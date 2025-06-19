import boto3
import re
import os
import json
import logging

logger = logging.getLogger()
logger.setLevel(logging.DEBUG)

logging.basicConfig(
    format='%(asctime)s - %(levelname)s - %(message)s'
)

sts_client = boto3.client('sts')
PACKAGE_MANAGER_TABLE = 'package-manager-dev'
ARCH_LIST = ["aarch64", "armv7l", "i386", "powerpc", "ppc64", "ppc64le", "s390x", "sparc", "universal", "x86_64"]
ASSUME_ROLE_ARN = "arn:aws:iam::712624343120:role/cross-account-s3-role"  

def extract_channel(object_key):
    logging.info(f"Extracting channel from object key: {object_key}")
    match = re.search(r'\b(stable|current)\b', object_key)
    if match:
        return match.group(1) 
    return None 

def convert_to_dynamodb_format(data):
    logging.info("Converting data to DynamoDB format")
    if isinstance(data, dict):
        return {k: {"M": convert_to_dynamodb_format(v)} if isinstance(v, dict) else {"S": str(v)} for k, v in data.items()}
    return {'S': str(data)}

def lambda_handler(event, context):
    try:
        # Log raw event for debugging
        logging.info("Raw event: %s", event)

        # Assume the role
        assumed_role = sts_client.assume_role(
            RoleArn=ASSUME_ROLE_ARN,
            RoleSessionName="CrossAccountAccessSession"
        )

        logging.info("Assumed role successfully.")
    
        s3_client = boto3.client(
            's3',
            aws_access_key_id=assumed_role['Credentials']['AccessKeyId'],
            aws_secret_access_key=assumed_role['Credentials']['SecretAccessKey'],
            aws_session_token=assumed_role['Credentials']['SessionToken']
        )
        
        dynamodb_client = boto3.client('dynamodb')
        logging.info("Lambda function started processing S3 bucket objects.")
        
        try:
            logging.info("Received event: %s", json.dumps(event, indent=2))
        except Exception as e:
            logging.error("Failed to format event as JSON: %s", str(e))

        bucket_name = event['Records'][0]['s3']['bucket']['name']
        object_key = event['Records'][0]['s3']['object']['key']

        logging.info(f"Bucket name: {bucket_name}")
        logging.info(f"Object key: {object_key}")

        channel = extract_channel(object_key)


        METADATA_TABLE = f"package-details-{channel}-dev"
        logging.info(f"Metadata table: {METADATA_TABLE}")
        
        if object_key.endswith('metadata.json'):
            pversion_folder = os.path.dirname(object_key) + '/'
        else :
            logging.info(f"metadata.json not created or updated: {object_key}")
            exit(0)
        objects = s3_client.list_objects_v2(Bucket=bucket_name, Prefix=pversion_folder)
        logging.info(f"Objects in bucket {bucket_name}: {objects}")
        if 'Contents' not in objects:
            logging.info(f"No objects found in the bucket: {bucket_name} under channel: {channel}")
            exit(0)
        
        
        for obj in objects['Contents']:
            object_key = obj['Key']
            if object_key.endswith('metadata.json'):
                parts = object_key.split('/')
                object_channel = extract_channel(parts[0])
                if len(parts) < 3 or object_channel != channel:
                    continue
                logging.info(f"Processing object: {object_key}")
                
                response = s3_client.get_object(Bucket=bucket_name, Key=object_key)
                file_content = response['Body'].read().decode('utf-8')

                logging.info(f"File content: {file_content}") 

                # Validate JSON parsing
                try:
                    json_content = json.loads(file_content)
                except json.JSONDecodeError as e:
                    logging.error("Failed to parse JSON file: %s", str(e))
                    raise              

                for product_name, product_versions in json_content.items():
                    logging.info(f"Processing product: {product_versions}")
                    for version, version_data in product_versions.items():
                        filtered_metadata = {key: value for key, value in version_data.items() if key != 'product-version-metadata'}
                        logging.info(f"Filtered metadata for {product_name} version {version}: {filtered_metadata}")
                        dynamodb_client.put_item(
                            TableName=METADATA_TABLE,
                            Item={
                                'product': {'S': product_name},
                                'version': {'S': version},
                                'metadata': {'M': convert_to_dynamodb_format(filtered_metadata)}
                            }
                        )
                
                logging.info("Processing package manager data")
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

        return {
            'statusCode': 200,
            'body': json.dumps('Successfully processed metadata files.')
        }
    
    except Exception as e:
        logging.error(f"Error: {str(e)}")
        return {
            'statusCode': 500,
            'body': json.dumps(f"Error: {str(e)}")
        }