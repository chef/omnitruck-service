import pandas as pd
import os
import boto3


def create_session():
    session = boto3.Session(
        aws_access_key_id=os.environ.get("ACCESS_KEY"),
        aws_secret_access_key=os.environ.get("SECRET_KEY"),
        region_name=os.environ.get("REGION")
    )
    return session


def filter_rows(df, product_name, sku_name):
    try:
        if product_name in df.columns and sku_name in df.columns:
            filtered_rows = []
            for index, row in df.iterrows():
                product = row[product_name]
                sku = row[sku_name]
                if pd.notna(sku) and sku != "":
                    filtered_rows.append(product)
            return filtered_rows
        else:
            return "One or both of the specified columns not found in the Excel sheet."
    except Exception as e:
        return f"An error occurred: {str(e)}"


def get_Columns(df):
    try:
        columns = df.columns.tolist()
        response_dict = {}
        for i, j in enumerate(columns[1:]):
            if "Content" in j or "UI" in j:
                continue
            if j == "Status":
                continue
            response = filter_rows(df, columns[0], j)
            response_dict[j] = response
        return response_dict
    except Exception as e:
        return f"An error occurred: {str(e)}"


def getRelated(df, value):
    response_dist = {}
    for software in value:
        ui_name = df.loc[df['Software'] == software,
                     'UI - Presentable Software Name'].values[0]
        software = software.split()
        if software:
            software = software[0]
        software = software.lower()
        response_dist[software] = ui_name
    return response_dist


def push_to_database(filepath):
    if os.path.exists(filepath) == False:
        return f"The given filepath doesnot exist: {str(filepath)}"
    df = pd.read_excel(filepath, sheet_name=0,
                       skiprows=1)
    data = get_Columns(df)
    session = create_session()
    table_name = os.getenv('RELATED_PRODUCTS_TABLE_NAME')
    try:
        dynamodb = session.resource('dynamodb')
        table = dynamodb.Table(table_name)
        for key, value in data.items():
            response = getRelated(df, value)
            item = {
                "bom": key,
                "products": response
            }
            table.put_item(Item=item)
        return "Data pushed successfully!"
    except Exception as e:
        return f"An error occurred: {str(e)}"


filename = input("Enter the filename: ")
response = push_to_database(filename)
print(response)
