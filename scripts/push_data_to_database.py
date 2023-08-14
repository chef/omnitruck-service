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


def filter_rows(file_path, product_name, sku_name):
    try:
        df = pd.read_excel(file_path)
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


def get_Columns(file_path):
    try:
        df = pd.read_excel(file_path)
        columns = df.columns.tolist()
        response_dict = {}
        for i, j in enumerate(columns[1:]):
            response = filter_rows(file_path, columns[0], j)
            if "-" in j:
                j = j.replace(" - ", "-").replace(" -", "-").replace("- ","-")
            sku_product = j.replace(" ", "-").lower()
            response_dict[sku_product] = response
        return response_dict
    except Exception as e:
        return f"An error occurred: {str(e)}"


def push_to_database(filepath):
    data = get_Columns(filepath)
    session = create_session()
    table_name = os.getenv('SKU_TABLE_NAME')
    try:
        dynamodb = session.resource('dynamodb')
        table = dynamodb.Table(table_name)
        for key, value in data.items():
            item = {
                "sku": key,
                "products": value
            }
            table.put_item(Item=item)
        return "Data pushed successfully!"
    except Exception as e:
        return f"An error occurred: {str(e)}"

filename = input("Enter the filename: ")
response = push_to_database(filename)
print(response)
