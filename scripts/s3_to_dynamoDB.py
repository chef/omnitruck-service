import boto3
import json
import logging

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)

METADATA_TABLE = 'package-details-dev'
PACKAGE_MANAGER_TABLE = 'package-manager-dev'

def convert_to_dynamodb_format(data):
    if isinstance(data, dict):
        return {k: {'M': convert_to_dynamodb_format(v)} if isinstance(v, dict) else {'S': str(v)} for k, v in data.items()}
    return {'S': str(data)}

def lambda_handler(event, context):
    s3_client = boto3.client('s3')
    dynamodb_client = boto3.client('dynamodb')
    
    try:
        bucket_name = event['bucket_name']
        logging.info(f"Bucket Name: {bucket_name}")
        
        objects = s3_client.list_objects_v2(Bucket=bucket_name)
        if 'Contents' not in objects:
            raise ValueError(f"No objects found in the bucket: {bucket_name}")
        
        for obj in objects['Contents']:
            object_key = obj['Key']
            
            if object_key.endswith('metadata.json'):
                logging.info(f"Processing S3 object: {object_key}")
                parts = object_key.split('/')
                if len(parts) < 3:
                    logging.warning(f"Skipping object with unexpected key structure: {object_key}")
                    continue
                
                response = s3_client.get_object(Bucket=bucket_name, Key=object_key)
                file_content = response['Body'].read().decode('utf-8')
                
                json_content = json.loads(file_content)
                
                stable_data = json_content.get('stable', {})
                for product_name, product_versions in stable_data.items():
                    
                    for version, version_data in product_versions.items():
                        
                        linux_contents = version_data.get('linux', {})
                        windows_contents = version_data.get('windows', {})

                        linux_dynamodb_format = convert_to_dynamodb_format(linux_contents)
                        windows_dynamodb_format = convert_to_dynamodb_format(windows_contents)

                        logging.info(f"Adding metadata for product: {product_name}, version: {version}")
                        dynamodb_client.put_item(
                            TableName=METADATA_TABLE,
                            Item={
                                'product': {'S': product_name},
                                'version': {'S': version},
                                'metadata': {'M': {
                                    'linux': {'M': linux_dynamodb_format},
                                    'windows': {'M': windows_dynamodb_format}
                                }}
                            }
                        )
                        
                        if linux_contents:
                            for arch, arch_data in linux_contents.items():
                                if isinstance(arch_data, dict):
                                    for package_type, package_details in arch_data.items():
                                        logging.info(f"Adding linux package: {package_type}")
                                        dynamodb_client.put_item(
                                            TableName=PACKAGE_MANAGER_TABLE,
                                            Item={
                                                'packages': {'S': package_type},
                                            }
                                        )
                                else:
                                    logging.warning(f"Unexpected data format for Linux architecture {arch}: {arch_data}")
                        else:
                            logging.info("No Linux content found.")

                        if windows_contents:
                            for arch, arch_data in windows_contents.items():
                                if isinstance(arch_data, dict):
                                    for package_type, package_details in arch_data.items():
                                        logging.info(f"Adding windows package: {package_type}")
                                        dynamodb_client.put_item(
                                            TableName=PACKAGE_MANAGER_TABLE,
                                            Item={
                                                'packages': {'S': package_type},
                                            }
                                        )
                                else:
                                    logging.warning(f"Unexpected data format for Windows architecture {arch}: {arch_data}")
                        else:
                            logging.info("No Windows content found.")

        logging.info("All relevant metadata.json files processed successfully.")
        
        return {
            'statusCode': 200,
            'body': json.dumps('Successfully processed metadata.json file.')
        }
    
    except Exception as e:
        logging.error(f"Error while running the lambda script: {str(e)}")
        return {
            'statusCode': 500,
            'body': json.dumps(f"Error: {str(e)}")
        }