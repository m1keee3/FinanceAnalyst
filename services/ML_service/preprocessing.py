# -*- coding: utf-8 -*-
"""
Preprocessing pipeline to replicate Alsulmi & Al-Shahrani (2022) data prep:
- Labeling with future 8-day average close (UP/DOWN)
- 31+ technical indicators per paper's Table 3
- Convert features to % daily changes
- Min–Max scaling to [0, 1]
References (methodology): labeling window n=8; technical indicators; normalization of daily changes then min–max scale.
"""

from __future__ import annotations
import os
import warnings
from dataclasses import dataclass
from typing import List, Tuple, Optional, Dict

import numpy as np
import pandas as pd
from sklearn.preprocessing import MinMaxScaler
from ta.momentum import RSIIndicator, StochasticOscillator, ROCIndicator, WilliamsRIndicator
from ta.trend import SMAIndicator, EMAIndicator, MACD, CCIIndicator, ADXIndicator
from ta.volatility import BollingerBands, AverageTrueRange
from ta.volume import OnBalanceVolumeIndicator, ChaikinMoneyFlowIndicator, AccDistIndexIndicator
from pathlib import Path
# Note: ta has PPO via MACD percentage? We'll compute PPO as MACD/EMA26 *100 ourselves.

warnings.filterwarnings("ignore")

# ---------------------- Settings ----------------------

# Входной файл с котировками
INPUT_PATH = Path("services\ML_service\data\SBER_daily_10y.csv")        # путь к твоему CSV
# Выходной файл
OUTPUT_PATH = Path("services\ML_service\data\preprocessed\preprocessed_SBER.csv") # куда сохранить результат

# Если нужно разделить train/test по времени
FIT_ON_START_DATE = None  # дата границы, None если не нужно
SCALE_PER_TICKER = False          # True если хочешь min-max отдельно по тикеру


# ---------------------------- Config ----------------------------

@dataclass
class PrepConfig:
    future_window: int = 8             # n for labeling (UP/DOWN)
    pct_cols_exclude: Tuple[str, ...] = ("label", "date", "ticker")
    # windows for multi-scale indicators (as in the paper they expanded several indicators over many cutoffs)
    ma_windows: Tuple[int, ...] = (5, 10, 15, 20, 25, 50, 100)
    osc_windows: Tuple[int, ...] = (5, 10, 14)  # e.g., RSI 14, WR 14, Stoch 14
    atr_windows: Tuple[int, ...] = (14,)
    boll_windows: Tuple[int, ...] = (20,)
    adx_windows: Tuple[int, ...] = (14,)
    aroon_windows: Tuple[int, ...] = (25,)      # paper’s Aroon uses 25
    clip_inf: bool = True
    dropna_after: bool = True
    scale_per_ticker: bool = False             # paper applies global min–max to [0,1]; keep False to fit on all data
    # You can provide an explicit train mask to fit scaler only on train period
    fit_on_mask_col: Optional[str] = None      # e.g., "is_train" boolean column
    random_state: int = 42


# ---------------------------- Utilities ----------------------------
def _aroon_up_down(close: pd.Series, window: int) -> Tuple[pd.Series, pd.Series]:
    """
    Aroon Up/Down:
      up   = 100 * (window - periods_since_max) / window
      down = 100 * (window - periods_since_min) / window
    """
    # индексы максимумов/минимумов за окна
    roll = close.rolling(window, min_periods=window)

    # position of last max/min inside window (0..window-1), считаем через argmax/argmin на скользящем массиве
    # используем трюк с shift для выравнивания
    def last_idx_of_max(a: pd.Series) -> pd.Series:
        vals = a.to_numpy()
        out = np.full_like(vals, fill_value=np.nan, dtype=float)
        buf = np.empty(window)
        for i in range(window-1, len(vals)):
            buf[:] = vals[i-window+1:i+1]
            out[i] = window - 1 - np.argmax(buf)  # сколько дней назад был максимум
        return pd.Series(out, index=a.index)

    def last_idx_of_min(a: pd.Series) -> pd.Series:
        vals = a.to_numpy()
        out = np.full_like(vals, fill_value=np.nan, dtype=float)
        buf = np.empty(window)
        for i in range(window-1, len(vals)):
            buf[:] = vals[i-window+1:i+1]
            out[i] = window - 1 - np.argmin(buf)
        return pd.Series(out, index=a.index)

    since_max = last_idx_of_max(close)
    since_min = last_idx_of_min(close)
    aroon_up = 100.0 * (window - since_max) / window
    aroon_down = 100.0 * (window - since_min) / window
    return aroon_up, aroon_down

def _ensure_types(df: pd.DataFrame) -> pd.DataFrame:
    # enforce dtypes
    df = df.copy()
    df["ticker"] = df["ticker"].astype(str)
    # try parse 'date' whether int like 20180101 or ISO
    if np.issubdtype(df["date"].dtype, np.number):
        df["date"] = pd.to_datetime(df["date"].astype(str), format="%Y%m%d", errors="coerce")
    else:
        df["date"] = pd.to_datetime(df["date"], errors="coerce")
    num_cols = ["open", "high", "low", "close", "volume"]
    for c in num_cols:
        df[c] = pd.to_numeric(df[c], errors="coerce")
    df = df.sort_values(["ticker", "date"])
    return df


def _future_avg_close_labeling(g: pd.DataFrame, n: int) -> pd.Series:
    # future average of next n closes (shift(-1).rolling(n).mean())
    future_avg = g["close"].shift(-1).rolling(n, min_periods=n).mean()
    label = (future_avg > g["close"]).astype(int)  # 1=UP, 0=DOWN
    return label


def _prev_direction(g: pd.DataFrame) -> pd.Series:
    # previous day's direction up/down (1/0)
    return (g["close"].pct_change() > 0).astype(int).shift(1)


def _typical_price(g: pd.DataFrame) -> pd.Series:
    return (g["high"] + g["low"] + g["close"]) / 3.0


def _weighted_close(g: pd.DataFrame) -> pd.Series:
    return (2 * g["close"] + g["high"] + g["low"]) / 4.0


def _true_range(g: pd.DataFrame) -> pd.Series:
    prev_close = g["close"].shift(1)
    tr = pd.concat([
        g["high"] - g["low"],
        (g["high"] - prev_close).abs(),
        (g["low"] - prev_close).abs()
    ], axis=1).max(axis=1)
    return tr


def _percentage_price_oscillator(close: pd.Series, fast: int = 12, slow: int = 26) -> pd.Series:
    # PPO = (EMA_fast - EMA_slow) / EMA_slow * 100
    ema_fast = close.ewm(span=fast, adjust=False, min_periods=fast).mean()
    ema_slow = close.ewm(span=slow, adjust=False, min_periods=slow).mean()
    ppo = (ema_fast - ema_slow) / ema_slow * 100.0
    return ppo


def _double_ema(close: pd.Series, span: int) -> pd.Series:
    ema = close.ewm(span=span, adjust=False, min_periods=span).mean()
    dema = 2 * ema - ema.ewm(span=span, adjust=False, min_periods=span).mean()
    return dema


def _triple_ema(close: pd.Series, span: int) -> pd.Series:
    ema1 = close.ewm(span=span, adjust=False, min_periods=span).mean()
    ema2 = ema1.ewm(span=span, adjust=False, min_periods=span).mean()
    ema3 = ema2.ewm(span=span, adjust=False, min_periods=span).mean()
    tema = 3 * (ema1 - ema2) + ema3
    return tema


def _triangular_ma(close: pd.Series, period: int) -> pd.Series:
    # TRIMA via double SMA
    sma1 = close.rolling(period, min_periods=period).mean()
    sma2 = sma1.rolling(period, min_periods=period).mean()
    return sma2


def _clip_inf_df(df: pd.DataFrame) -> pd.DataFrame:
    df = df.replace([np.inf, -np.inf], np.nan)
    return df


# ---------------------------- Feature Engineering ----------------------------

def build_features_for_group(g: pd.DataFrame, cfg: PrepConfig) -> pd.DataFrame:
    g = g.copy()
    # Base stats already present: open, high, low, close, volume
    # Previous day's direction (binary)
    g["prev_dir"] = _prev_direction(g)

    # Multi-window moving averages
    for w in cfg.ma_windows:
        g[f"sma_{w}"] = SMAIndicator(close=g["close"], window=w, fillna=False).sma_indicator()
        g[f"ema_{w}"] = EMAIndicator(close=g["close"], window=w, fillna=False).ema_indicator()
        g[f"dema_{w}"] = _double_ema(g["close"], w)
        g[f"tema_{w}"] = _triple_ema(g["close"], w)
        g[f"trima_{w}"] = _triangular_ma(g["close"], w)

    # MACD classic (12,26,9)
    macd = MACD(close=g["close"], window_slow=26, window_fast=12, window_sign=9, fillna=False)
    g["macd"] = macd.macd()
    g["macd_signal"] = macd.macd_signal()
    g["macd_hist"] = macd.macd_diff()

    # Typical price and Weighted close
    g["tp"] = _typical_price(g)
    g["wc"] = _weighted_close(g)

    # Williams %R and Stochastic Oscillator, Momentum/ROC
    for w in cfg.osc_windows:
        try:
            g[f"wr_{w}"] = WilliamsRIndicator(high=g["high"], low=g["low"], close=g["close"], lbp=w, fillna=False).williams_r()
        except Exception:
            g[f"wr_{w}"] = np.nan
        so = StochasticOscillator(high=g["high"], low=g["low"], close=g["close"], window=w, smooth_window=3, fillna=False)
        g[f"stoch_k_{w}"] = so.stoch()
        g[f"stoch_d_{w}"] = so.stoch_signal()
        g[f"roc_{w}"] = ROCIndicator(close=g["close"], window=w).roc()

    # CCI, RSI, ADX
    for w in cfg.osc_windows:
        g[f"cci_{w}"] = CCIIndicator(high=g["high"], low=g["low"], close=g["close"], window=w, constant=0.015).cci()
    g["rsi_14"] = RSIIndicator(close=g["close"], window=14).rsi()
    for w in cfg.adx_windows:
        adx = ADXIndicator(high=g["high"], low=g["low"], close=g["close"], window=w)
        g[f"adx_{w}"] = adx.adx()

    # PPO (percentage price oscillator)
    g["ppo"] = _percentage_price_oscillator(g["close"], fast=12, slow=26)

    # True Range and ATR
    g["tr"] = _true_range(g)
    for w in cfg.atr_windows:
        g[f"atr_{w}"] = AverageTrueRange(high=g["high"], low=g["low"], close=g["close"], window=w).average_true_range()

    # Aroon up/down (period=25)
    for w in cfg.aroon_windows:
        up, down = _aroon_up_down(g["close"], w)
        g[f"aroon_up_{w}"] = up
        g[f"aroon_down_{w}"] = down

    # OBV, Money Flow (we’ll use Chaikin Money Flow; paper also mentions Acc/Dist and Balance of Power)
    g["obv"] = OnBalanceVolumeIndicator(close=g["close"], volume=g["volume"]).on_balance_volume()
    try:
        g["cmf_20"] = ChaikinMoneyFlowIndicator(high=g["high"], low=g["low"], close=g["close"], volume=g["volume"], window=20).chaikin_money_flow()
    except Exception:
        g["cmf_20"] = np.nan
    # Accumulation/Distribution Index (Chaikin ADL)
    try:
        g["chaikin_ad"] = AccDistIndexIndicator(high=g["high"], low=g["low"], close=g["close"], volume=g["volume"]).acc_dist_index()
    except Exception:
        g["chaikin_ad"] = np.nan

    # Balance of Power (approx)
    rng = (g["high"] - g["low"]).replace(0, np.nan)
    g["bop"] = (g["close"] - g["open"]) / rng

    return g


def add_labels(df: pd.DataFrame, cfg: PrepConfig) -> pd.DataFrame:
    df = df.copy()

    def label_series(s: pd.Series, n: int) -> pd.Series:
        # среднее следующих n закрытий
        fut = s.shift(-1).rolling(n, min_periods=n).mean()
        return (fut > s).astype(int)

    # возвращаем ровно серию, потом выравниваем индекс и кладём в колонку
    lab = (df.groupby("ticker")["close"]
             .apply(lambda s: label_series(s, cfg.future_window))
             .reset_index(level=0, drop=True))
    df["label"] = lab
    return df

def to_daily_pct_changes(df: pd.DataFrame, cfg: PrepConfig) -> pd.DataFrame:
    df = df.copy()
    # convert all numeric features to pct_change within each ticker, except excluded
    numeric_cols = df.select_dtypes(include=[np.number]).columns.tolist()
    # keep label numeric, but we do not pct_change it
    pct_cols = [c for c in numeric_cols if c not in cfg.pct_cols_exclude and c != "label"]
    def pct_block(g: pd.DataFrame) -> pd.DataFrame:
        g_pct = g[pct_cols].pct_change()
        # keep non-pct columns as is
        keep = g.drop(columns=pct_cols)
        return pd.concat([keep, g_pct], axis=1)
    df = df.groupby("ticker", group_keys=False).apply(pct_block)
    return df


def minmax_scale(df: pd.DataFrame, cfg: PrepConfig) -> Tuple[pd.DataFrame, Dict[str, MinMaxScaler]]:
    df = df.copy()
    scalers: Dict[str, MinMaxScaler] = {}
    # scale all numeric features except excluded and label
    numeric_cols = df.select_dtypes(include=[np.number]).columns.tolist()
    feat_cols = [c for c in numeric_cols if c not in cfg.pct_cols_exclude and c != "label"]

    if cfg.fit_on_mask_col and cfg.fit_on_mask_col in df.columns:
        fit_mask = df[cfg.fit_on_mask_col].astype(bool).values
    else:
        fit_mask = np.ones(len(df), dtype=bool)

    if cfg.scale_per_ticker:
        out = []
        for tkr, g in df.groupby("ticker"):
            scaler = MinMaxScaler()
            fit_idx = g.index[fit_mask[g.index]]
            scaler.fit(g.loc[fit_idx, feat_cols].values)
            g_scaled = g.copy()
            g_scaled.loc[:, feat_cols] = scaler.transform(g[feat_cols].values)
            out.append(g_scaled)
            scalers[str(tkr)] = scaler
        df_scaled = pd.concat(out, axis=0).sort_values(["ticker", "date"])
    else:
        scaler = MinMaxScaler()
        scaler.fit(df.loc[fit_mask, feat_cols].values)
        df.loc[:, feat_cols] = scaler.transform(df[feat_cols].values)
        df_scaled = df
        scalers["__global__"] = scaler

    return df_scaled, scalers


# ---------------------------- Main API ----------------------------

def preprocess_quotes(
    quotes: pd.DataFrame,
    cfg: Optional[PrepConfig] = None
) -> Tuple[pd.DataFrame, Dict[str, MinMaxScaler]]:
    """
    Input columns (per paper): ticker, date, open, high, low, close, volume
    Returns feature table with label, daily % changes, and min–max scaled features in [0,1].
    """
    cfg = cfg or PrepConfig()
    df = _ensure_types(quotes)

    # Labeling with future average close (n=8)
    df = add_labels(df, cfg)

    # Build features per ticker
    df = df.groupby("ticker", group_keys=False).apply(build_features_for_group, cfg=cfg)

    # Percent changes (daily) for ALL features (as in the paper: learn on changes, not absolutes)
    df = to_daily_pct_changes(df, cfg)

    # Clean infinities
    if cfg.clip_inf:
        df = _clip_inf_df(df)

    # Drop rows with NaNs produced by rolling/pct_change/shift
    if cfg.dropna_after:
        # at least drop where label is NaN or close pct is NaN
        df = df.dropna(subset=["label"])
        # optional: drop any remaining NaNs in features
        feat_null_cols = [c for c in df.columns if c not in ("ticker", "date")]
        df = df.dropna(subset=feat_null_cols)

    # Min–Max scaling to [0,1]
    df_scaled, scalers = minmax_scale(df, cfg)

    # Final sort
    df_scaled = df_scaled.sort_values(["ticker", "date"]).reset_index(drop=True)
    return df_scaled, scalers


# ---------------------------- CLI example ----------------------------

def main():
    print("=== Загрузка данных ===")
    raw = pd.read_csv(INPUT_PATH)
    raw = _ensure_types(raw)

    cfg = PrepConfig()
    if FIT_ON_START_DATE is not None:
        raw["is_train"] = raw["date"] < FIT_ON_START_DATE
        cfg.fit_on_mask_col = "is_train"
    cfg.scale_per_ticker = SCALE_PER_TICKER

    print(f"Обрабатываем {raw['ticker'].nunique()} тикеров, {len(raw)} строк...")

    df_prepped, scalers = preprocess_quotes(raw, cfg)

    OUTPUT_PATH.parent.mkdir(parents=True, exist_ok=True)
    df_prepped.to_csv(OUTPUT_PATH, index=False)
    print(f"✅ Сохранено: {OUTPUT_PATH}")
    print(f"Размер: {df_prepped.shape[0]} строк, {df_prepped.shape[1]} колонок")
    print(f"Кол-во классов:\n{df_prepped['label'].value_counts(normalize=True)}")

if __name__ == "__main__":
    main()
