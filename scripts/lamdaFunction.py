import json
import boto3
import urllib.parse
import urllib.request

s3_client = boto3.client("s3")

DYNAMO_TABLE = 'metadata-production'

def lambda_handler(event, context):
   if "manualTrigger" in event:
      bucket = event["s3Bucket"]
      addOldHabData(getAllReleasedHabVersions(), bucket)
   else:
      bucket = event['Records'][0]['s3']['bucket']['name']
      fileupdated = event['Records'][0]['s3']['object']['key']
      key = fileupdated[0:fileupdated.rindex("/")+1]
      if "automate" in key:
         shaFiles = readAllFiles(bucket, key, "automate")
         shaContent = readSHAValue(bucket, shaFiles)
         fetchPlatformInfoAutomate(shaContent, "latest")
      elif "hab" in key:
         shaFiles = readAllFiles(bucket, key, "hab")
         shaContent = readSHAValue(bucket, shaFiles)
         version = readVersion(bucket, key+"manifest.json")
         fetchPlatformInfoHab(shaContent, version)
   
def readAllFiles(bucket, key, setup):
   files_resp = s3_client.list_objects_v2(Bucket=bucket, Prefix=key, StartAfter=key,)
   files = files_resp["Contents"]
   shaFiles = []
   for file in files:
      if ".txt" in file["Key"] or "documentation" in file["Key"] or ".asc" in file["Key"] or "manifest" in file["Key"]:
         continue
      elif ".sha256sum" in file["Key"]:
         if setup == "automate" and file["LastModified"].year < 2023:
            continue
         shaFiles.append(file["Key"])
   return shaFiles

def readVersion(bucket, file):
   file_content = s3_client.get_object(Bucket=bucket, Key=file)["Body"].read().decode('utf-8')
   body = json.loads(file_content)
   return json.loads(file_content)["version"]

def readSHAValue(bucket, shaFiles):
   shaContent = []
   for file in shaFiles:
      file_content = s3_client.get_object(Bucket=bucket, Key=file)["Body"].read().decode('utf-8')
      shaContent.append(file_content)
   return shaContent
   
def fetchPlatformInfoHab(shaContent, version):
   productData = {"metadata": []}
   productData["product"] = "habitat"
   productData["version"] = version
   for shaInfo in shaContent:
      archValue = shaInfo.split(" ")[-1].strip()
      archStruct = archValue.split("-")
      productData["metadata"].append({"platform": archStruct[1], "sha256": shaInfo.split(" ")[0], "architecture": archStruct[-1].split(".")[0], "filename": archValue })
   addDataToDynamo(productData)
   
def fetchPlatformInfoAutomate(shaContent, ver):
   productData = {"metadata": []}
   productData["product"] = "automate"
   productData["version"] = ver
   for shaInfo in shaContent:
      archValue = shaInfo.split(" ")[-1].strip()
      archStruct = archValue.split("_")
      productData["metadata"].append({"platform": archStruct[1], "sha256": shaInfo.split(" ")[0], "architecture": archStruct[-1].split(".")[0], "filename": archValue })
   addDataToDynamo(productData)
   
def getReleasedHabVersionsAPI(range):
   url = "https://bldr.habitat.sh/v1/depot/channels/core/stable/pkgs/hab?range="+range
   releasedVersion = []
   versions = json.loads(urllib.request.urlopen(urllib.request.Request(url=url,method='GET')).read())
   for version in versions["data"]:
      releasedVersion.append(version["version"])
   return releasedVersion

def getAllReleasedHabVersions():
   range = 0
   allReleasedVersions = []
   while True:
      versions = getReleasedHabVersionsAPI(str(range))
      if len(versions) < 50:
         allReleasedVersions += list(set(versions))
         return allReleasedVersions
      range = range + len(versions)
      allReleasedVersions += list(set(versions))

def addOldHabData(versions, bucket):
   for ver in versions:
      key = "files/habitat/"+ver+"/"
      file_content = s3_client.list_objects_v2(Bucket=bucket, Prefix=key, StartAfter=key)
      if "Contents" in file_content:
         shaFiles = readAllFiles(bucket, key, "hab")
         shaContent = readSHAValue(bucket, shaFiles)
         version = readVersion(bucket, key+"manifest.json")
         fetchPlatformInfoHab(shaContent, version)

def addDataToDynamo(productData):
   dynamo_client = boto3.resource("dynamodb")
   table = dynamo_client.Table(DYNAMO_TABLE)
   response = table.put_item(Item = productData)
   print(response)
   