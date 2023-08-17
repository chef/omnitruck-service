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
        return f"An error occurred while filtering the columns: {str(e)}"


def get_Columns(file_path):
    try:
        df = pd.read_excel(file_path)
        columns = df.columns.tolist()
        response_dict = {}
        for i, j in enumerate(columns[1:]):
            response = filter_rows(df, columns[0], j)
            response_dict[j] = response
        return response_dict
    except Exception as e:
        return f"An error occurred: {str(e)}"


def push_to_database(filepath):
    if os.path.exists(filepath) == False:
        return f"The given filepath doesnot exist: {str(filepath)}"
    data = get_Columns(filepath)
    session = create_session()
    table_name = os.getenv('RELATED_PRODUCTS_TABLE_NAME')
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
