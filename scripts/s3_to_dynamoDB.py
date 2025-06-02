import boto3
import json

METADATA_TABLE = 'metadata'
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
        objects = s3_client.list_objects_v2(Bucket=bucket_name)
        if 'Contents' not in objects:
            raise ValueError(f"No objects found in the bucket: {bucket_name}")
        
        for obj in objects['Contents']:
            object_key = obj['Key']
            
            if object_key.endswith('metadata.json'):
                parts = object_key.split('/')
                if len(parts) < 3:
                    print(f"Skipping object with unexpected key structure: {object_key}")
                    continue
                
                response = s3_client.get_object(Bucket=bucket_name, Key=object_key)
                file_content = response['Body'].read().decode('utf-8')
                
                # Convert file_content to JSON
                json_content = json.loads(file_content)
                
                # Navigate to the stable folder dynamically
                stable_data = json_content.get('stable', {})
                for product_name, product_versions in stable_data.items():
                    
                    for version, version_data in product_versions.items():
                        
                        linux_contents = version_data.get('linux', {})
                        windows_contents = version_data.get('windows', {})

                        linux_dynamodb_format = convert_to_dynamodb_format(linux_contents)
                        windows_dynamodb_format = convert_to_dynamodb_format(windows_contents)


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
                                        dynamodb_client.put_item(
                                            TableName=PACKAGE_MANAGER_TABLE,
                                            Item={
                                                'packages': {'S': package_type},
                                            }
                                        )
                                else:
                                    print(f"Unexpected data format for Linux architecture {arch}: {arch_data}")
                        else:
                            print("No Linux content found.")

                        if windows_contents:
                            for arch, arch_data in windows_contents.items():
                                if isinstance(arch_data, dict):
                                    for package_type, package_details in arch_data.items():
                                        dynamodb_client.put_item(
                                            TableName=PACKAGE_MANAGER_TABLE,
                                            Item={
                                                'packages': {'S': package_type},
                                            }
                                        )
                                else:
                                    print(f"Unexpected data format for Windows architecture {arch}: {arch_data}")
                        else:
                            print("No Windows content found.")

        print("All relevant metadata.json files processed successfully.")
        
        return {
            'statusCode': 200,
            'body': json.dumps('Successfully processed metadata.json file.')
        }
    
    except Exception as e:
        print(f"Error: {str(e)}")
        return {
            'statusCode': 500,
            'body': json.dumps(f"Error: {str(e)}")
        }