import sys
import requests
import pandas as pd
import plotly.graph_objects as go

def fetch_moex_data(ticker, date_from, date_till):
    url = f"https://iss.moex.com/iss/history/engines/stock/markets/shares/securities/{ticker}.json"
    params = {'from': date_from, 'till': date_till}
    response = requests.get(url, params=params)
    response.raise_for_status()

    columns = response.json()['history']['columns']
    rows = response.json()['history']['data']

    df = pd.DataFrame(rows, columns=columns)
    df = df[df['BOARDID'] == 'TQBR']
    df['TRADEDATE'] = pd.to_datetime(df['TRADEDATE'])
    df.set_index('TRADEDATE', inplace=True)

    return df[['OPEN', 'HIGH', 'LOW', 'CLOSE']]

def plot_and_save_candlestick(df, ticker, file_path):
    fig = go.Figure(data=[go.Candlestick(
        x=df.index.strftime('%Y-%m-%d'),
        open=df['OPEN'],
        high=df['HIGH'],
        low=df['LOW'],
        close=df['CLOSE']
    )])

    fig.update_layout(
        title=f"{ticker} Candlestick Chart",
        xaxis_title="Date",
        yaxis_title="Price",
        xaxis_rangeslider_visible=False
    )

    # Сохранение в PNG
    fig.write_image(file_path, width=1920, height=1080, scale=2)

if __name__ == "__main__":
    from pathlib import Path

    ticker = sys.argv[1]
    date_from = sys.argv[2]
    date_till = sys.argv[3]
    dir_name = sys.argv[4]
    file_name = sys.argv[5]

    print(date_from, date_till)

    df = fetch_moex_data(ticker, date_from, date_till)

    desktop_path = Path.home() / "Desktop" / "Charts" / dir_name
    desktop_path.mkdir(parents=True, exist_ok=True)

    file_name += ".png"
    file_path = desktop_path / file_name

    plot_and_save_candlestick(df, ticker, str(file_path))
