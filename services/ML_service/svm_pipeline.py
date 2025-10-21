# -*- coding: utf-8 -*-
"""
SVM-пайплайн для предсказания label (UP/DOWN) на данных после препроцессинга.
Особенности:
- Временная разбивка (без утечек): TimeSeriesSplit
- Грид-поиск по C, gamma (RBF)
- class_weight='balanced' на случай перекоса
- Метрики: accuracy, F1, ROC-AUC, отчёт по классам, confusion matrix
- Сохранение модели и предиктов

Ожидаемый входной CSV: колонки 'ticker','date','label' + числовые фичи.
Дата должна быть монотонной (по всему набору), как в нашем препроцессинге.
"""

import os
from pathlib import Path
import numpy as np
import pandas as pd
import joblib
import matplotlib.pyplot as plt

from sklearn.model_selection import TimeSeriesSplit, GridSearchCV
from sklearn.metrics import accuracy_score, f1_score, roc_auc_score, classification_report, confusion_matrix
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import StandardScaler
from sklearn.svm import SVC

# ------------------- НАСТРОЙКИ -------------------

DATA_PATH = Path("services\ML_service\data\preprocessed\preprocessed_SBER.csv")          # твой датасет после препроцессинга
OUT_DIR   = Path("services/ML_service/artifacts")                  # куда складывать артефакты
N_SPLITS  = 5                                      # TimeSeriesSplit (expanding window)
TEST_SIZE_FRACTION = 0.2                           # доля последних данных под финальный тест

PARAM_GRID = {
    "clf__C": [0.3, 1.0, 3.0, 10.0],
    "clf__gamma": ["scale", 0.03, 0.01],
    "clf__kernel": ["rbf"],
}

RANDOM_STATE = 42

# ------------------- УТИЛИТЫ -------------------

def load_data(path: Path) -> pd.DataFrame:
    df = pd.read_csv(path)
    if "date" not in df.columns or "label" not in df.columns:
        raise ValueError("Нужны колонки 'date' и 'label'.")
    df["date"] = pd.to_datetime(df["date"], errors="coerce")
    df = df.dropna(subset=["date", "label"]).sort_values("date").reset_index(drop=True)
    # бинаризуем на всякий
    df["label"] = (df["label"] > 0).astype(int)
    return df

def select_features(df: pd.DataFrame) -> list[str]:
    num_cols = df.select_dtypes(include=[np.number]).columns.tolist()
    # Убираем служебные
    drop = {"label"}
    feat = [c for c in num_cols if c not in drop]
    return feat

def time_split_indices(n_rows: int, test_frac: float):
    """Возвращает индексы train и test по последней доле наблюдений."""
    test_len = max(1, int(n_rows * test_frac))
    train_end = n_rows - test_len
    train_idx = np.arange(0, train_end)
    test_idx = np.arange(train_end, n_rows)
    return train_idx, test_idx

def plot_confusion(cm: np.ndarray, labels=("0","1"), title="Confusion matrix"):
    plt.figure(figsize=(4,3))
    plt.imshow(cm, interpolation="nearest", aspect="auto")
    plt.title(title)
    plt.colorbar()
    tick_marks = np.arange(len(labels))
    plt.xticks(tick_marks, labels); plt.yticks(tick_marks, labels)
    # подписи
    thresh = cm.max() / 2.0
    for i in range(cm.shape[0]):
        for j in range(cm.shape[1]):
            plt.text(j, i, format(cm[i, j], "d"),
                     ha="center", va="center",
                     color="white" if cm[i, j] > thresh else "black")
    plt.xlabel("Predicted"); plt.ylabel("True")
    plt.tight_layout()
    plt.show()

# ------------------- ОСНОВНОЙ КОД -------------------

def main():
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    df = load_data(DATA_PATH)
    features = select_features(df)

    X_all = df[features].replace([np.inf, -np.inf], np.nan).dropna()
    # важно синхронизировать y и date
    aligned_idx = X_all.index
    y_all = df.loc[aligned_idx, "label"].astype(int).values
    dates_all = df.loc[aligned_idx, "date"].values

    # Финальный holdout-тест: последние TEST_SIZE_FRACTION наблюдений
    tr_idx, te_idx = time_split_indices(len(X_all), TEST_SIZE_FRACTION)
    X_train, y_train = X_all.iloc[tr_idx], y_all[tr_idx]
    X_test,  y_test  = X_all.iloc[te_idx],  y_all[te_idx]

    print(f"Обучение: {X_train.shape}, Тест: {X_test.shape}")
    print(f"Train dates: {pd.to_datetime(dates_all[tr_idx[0]]).date()} — {pd.to_datetime(dates_all[tr_idx[-1]]).date()}")
    print(f"Test  dates: {pd.to_datetime(dates_all[te_idx[0]]).date()} — {pd.to_datetime(dates_all[te_idx[-1]]).date()}")

    # Пайплайн: стандартная шкала + SVC
    # Да, у нас уже min–max [0,1], но StandardScaler часто помогает для ядровых методов.
    pipe = Pipeline([
        ("scaler", StandardScaler(with_mean=False)),  # with_mean=False безопаснее на разреженных/масштабируемых фичах
        ("clf", SVC(probability=True, class_weight="balanced", random_state=RANDOM_STATE))
    ])

    # Временная CV без утечек
    tscv = TimeSeriesSplit(n_splits=N_SPLITS)  # expanding window

    # Грид-поиск по AUC
    grid = GridSearchCV(
        estimator=pipe,
        param_grid=PARAM_GRID,
        scoring="roc_auc",
        cv=tscv,
        n_jobs=-1,
        verbose=1
    )
    grid.fit(X_train, y_train)

    print("\nЛучшие гиперы:", grid.best_params_)
    print("CV best AUC:", round(grid.best_score_, 4))

    # Обучаем лучшую модель на всём train
    best = grid.best_estimator_
    joblib.dump(best, OUT_DIR / "svm_model.joblib")

    # Оценка на отложенном тесте
    y_prob = best.predict_proba(X_test)[:, 1]
    y_pred = (y_prob >= 0.5).astype(int)

    acc = accuracy_score(y_test, y_pred)
    f1  = f1_score(y_test, y_pred)
    try:
        auc = roc_auc_score(y_test, y_prob)
    except ValueError:
        auc = float("nan")

    print("\n=== TEST METRICS ===")
    print("Accuracy:", round(acc, 4))
    print("F1:", round(f1, 4))
    print("ROC-AUC:", round(auc, 4))

    print("\nClassification report:")
    print(classification_report(y_test, y_pred, digits=4))

    cm = confusion_matrix(y_test, y_pred)
    print("Confusion matrix:\n", cm)
    plot_confusion(cm, labels=["DOWN(0)", "UP(1)"], title="Confusion (SVM)")

    # Сохраним предсказания
    out_pred = pd.DataFrame({
        "date": dates_all[te_idx],
        "y_true": y_test,
        "y_pred": y_pred,
        "p_up": y_prob
    })
    out_pred.to_csv(OUT_DIR / "test_predictions.csv", index=False)
    print(f"\nСохранено:\n- модель: {OUT_DIR / 'svm_model.joblib'}\n- предсказания: {OUT_DIR / 'test_predictions.csv'}")

    # Быстрый график средних вероятностей по времени
    fig = plt.figure(figsize=(10,3.5))
    plt.plot(out_pred["date"], out_pred["p_up"])
    plt.title("SVM: вероятность UP на тесте")
    plt.xlabel("date"); plt.ylabel("p_up")
    plt.grid(True)
    plt.tight_layout()
    plt.show()

if __name__ == "__main__":
    main()
