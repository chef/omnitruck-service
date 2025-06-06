import boto3
import json
import logging

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)

METADATA_TABLE = 'package-details-dev'
PACKAGE_MANAGER_TABLE = 'package-manager-dev'

ARCH_LIST = ["aarch64", "armv7l", "i386", "powerpc", "ppc64", "ppc64le", "s390x", "sparc", "universal", "x86_64"]

def convert_to_dynamodb_format(data):
    if isinstance(data, dict):
        return {k: {"M": convert_to_dynamodb_format(v)} if isinstance(v, dict) else {"S": str(v)} for k, v in data.items()}
    return {'S': str(data)}

def check_if_exists(dynamodb_client, product_name, product_version):
    logging.info(f"Checking if product {product_name} with version {product_version} exists in DynamoDB.")
    response = dynamodb_client.get_item(
        TableName=METADATA_TABLE,
        Key={
            'product': {'S': product_name},
            'version': {'S': product_version}
        }
    )
    return response.get('Item')

def lambda_handler(event, context):
    s3_client = boto3.client('s3')
    dynamodb_client = boto3.client('dynamodb')
    print("Lambda function started processing S3 bucket objects.")
    
    try:
        logging.info("Received event: %s", json.dumps(event, indent=2))
        bucket_name = event['Records'][0]['s3']['bucket']['name']
        object_key = event['Records'][0]['s3']['object']['key']
        
        parts = object_key.split('/')
        if len(parts) < 2:
            raise ValueError(f"Unexpected object key structure: {object_key}")
        channel = parts[0]
        
        logging.info(f"Bucket Name: {bucket_name}")
        logging.info(f"Channel: {channel}")
        
        # List objects in the bucket
        objects = s3_client.list_objects_v2(Bucket=bucket_name, Prefix=f"{channel}/")
        if 'Contents' not in objects:
            logging.error(f"No objects found in the bucket: {bucket_name} under channel: {channel}")
            raise ValueError(f"No objects found in the bucket: {bucket_name} under channel: {channel}")
        
        for obj in objects['Contents']:
            object_key = obj['Key']
            if object_key.endswith('metadata.json'):
                parts = object_key.split('/')
                if len(parts) < 3 or parts[0] != channel:
                    continue
                
                response = s3_client.get_object(Bucket=bucket_name, Key=object_key)
                file_content = response['Body'].read().decode('utf-8')
                
                logging.info(f"Processing file: {object_key}")
                json_content = json.loads(file_content)

                channel_data = json_content.get(channel, {})

                for product_name, product_versions in channel_data.items():
                    for version, version_data in product_versions.items():
                        filtered_metadata = {key: value for key, value in version_data.items() if key != 'product-version-metadata'}
                        existing_product = check_if_exists(dynamodb_client, product_name, version)
                        if existing_product:
                            logging.info(f"Product {product_name} with version {version} already exists in DynamoDB. Updating metadata.")
                            existing_metadata = existing_product['metadata']['M']
                            existing_metadata[channel] = {"M": convert_to_dynamodb_format(filtered_metadata)}
                            dynamodb_client.put_item(
                                TableName=METADATA_TABLE,
                                Item={
                                    'product': {'S': product_name},
                                    'version': {'S': version},
                                    'metadata': {'M': existing_metadata}
                                }
                            )
                            
                        else:
                            filtered_metadata = {key: value for key, value in version_data.items() if key != 'product-version-metadata'}
                            metadata = {
                                channel: filtered_metadata
                            }
                            dynamodb_client.put_item(
                                TableName=METADATA_TABLE,
                                Item={
                                    'product': {'S': product_name},
                                    'version': {'S': version},
                                    'metadata': {'M': convert_to_dynamodb_format(metadata)}
                                }
                            )
                
                logging.info(f"Processing package manager data")
                for platform, platform_data in version_data.items():
                    if isinstance(platform_data, dict):
                        for arch, arch_data in platform_data.items():
                            if isinstance(arch_data, dict):
                                for package_type, package_details in arch_data.items():
                                    if isinstance(package_details, dict):
                                        if package_type in ARCH_LIST:
                                            continue
                                        
                                        dynamodb_client.put_item(
                                            TableName=PACKAGE_MANAGER_TABLE,
                                            Item={
                                                'packages': {'S': package_type}
                                            }
                                        )
                            else:
                                logging.info(f"Unexpected data format for architecture {arch}: {arch_data}")
                    else:
                        logging.info(f"Unexpected data format for platform {platform}: {platform_data}")

                        logging.info("All relevant metadata.json files processed successfully.")
        
        return {
            'statusCode': 200,
            'body': json.dumps('Successfully processed metadata files.')
        }
    
    except Exception as e:
        print(f"Error: {str(e)}")
        return {
            'statusCode': 500,
            'body': json.dumps(f"Error: {str(e)}")
        }