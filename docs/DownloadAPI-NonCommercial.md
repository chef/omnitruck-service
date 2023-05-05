# Download API Documentation (Non-Commercial License)
 There are two environments available for Download API i.e: `production`, `acceptance`. Below are the base url for emvironments. `Acceptance` environment is an internal testing environment for Download API
 ### Base url for Acceptance environments.
 
 `trial` : https://trial-acceptance.downloads.chef.co/
 
 `opensource` : https://opensource-acceptance.downloads.chef.co/

 ### Base url for Production environments.
 
  `trial` : https://ChefDownload-Trial.chef.io
 
 `opensource` : https://ChefDownload-Community.chef.io
 
 
 ## Download API types

Download API has three different use cases : `trial`, `opensource`, and `commercial`. These use cases indicate the type of license which the user may have - 
`opensource`, `trial` , `commercial`. A user may be able to get a `commercial` license or `trial` or `free` license which will qualify them to use the 
product based on the license policies. The Download API for commercial licenses is discussed in the [link here]{https://github.com/chef/omnitruck-service/blob/main/docs/DownloadAPI-Commercial.md}.

 ### Download API - Trial

When a user is having a trial license they can connect to the `trial` instance of the API and use that to provide them with product information and also be able to download the trial version of the chef products for the which the license applies.

Valid `<CHANNEL>` values in endpoint URLs is limited to `stable` 
Endpoint results are limited to the most recent version of any product unless a valid commercial `license_id` is provided.

### Examples

### products
Returns a valid list of valid product keys.
```
curl -X 'GET' \
  'https://chefdownload-trial.chef.io/products?eol=false' \
  -H 'accept: application/json'
  
  [
  "chef",
  "chef-backend",
  "chef-server",
  "chef-workstation",
  "inspec",
  "manage",
  "supermarket"
]

curl -X 'GET' \
  'https://chefdownload-trial.chef.io/products?eol=true' \
  -H 'accept: application/json'
  
  [
  "analytics",
  "angry-omnibus-toolchain",
  "angrychef",
  "automate",
  "chef",
  "chef-foundation",
  "chef-universal",
  "chef-backend",
  "chef-server",
  "chef-server-ha-provisioning",
  "chef-workstation",
  "chefdk",
  "compliance",
  "delivery",
  "ha",
  "harmony",
  "inspec",
  "mac-bootstrapper",
  "manage",
  "marketplace",
  "omnibus-toolchain",
  "omnibus-gcc",
  "private-chef",
  "push-jobs-client",
  "push-jobs-server",
  "reporting",
  "supermarket",
  "sync"
]

```
### architectures
Returns a valid list of valid architectures for the chef products. Any of these architectures can be used in the m query string value in various endpoints below. 
```
curl -X 'GET' \
  'https://chefdownload-trial.chef.io/architectures' \
  -H 'accept: application/json'
  [
  "aarch64",
  "armv7l",
  "i386",
  "powerpc",
  "ppc64",
  "ppc64le",
  "s390x",
  "sparc",
  "universal",
  "x86_64"
]
```
### Platforms
Returns a valid list of valid platform keys along with full friendly names. Any of these platform keys can be used in the p query string value in various endpoints below.
```
curl -X 'GET' \
  'https://chefdownload-trial.chef.io/platforms' \
  -H 'accept: application/json'

{
  "aix": "AIX",
  "amazon": "Amazon Linux",
  "debian": "Debian GNU/Linux",
  "el": "Red Hat Enterprise Linux/CentOS",
  "freebsd": "FreeBSD",
  "ios_xr": "Cisco IOS-XR",
  "mac_os_x": "macOS",
  "nexus": "Cisco NX-OS",
  "sles": "SUSE Linux Enterprise Server",
  "solaris2": "Solaris",
  "suse": "openSUSE",
  "ubuntu": "Ubuntu Linux",
  "windows": "Windows"
}
```

### /\<CHANNEL>/\<PRODUCT>/versions/all
 
Get a list of all available version numbers for a particular channel and product combination 

### /\<CHANNEL>/\<PRODUCT>/versions/latest
Get the latest version number for a particular channel and product combination.

Example Request:
 
### /\<CHANNEL>/\<PRODUCT>/packages
Get the full list of all packages for a particular channel and product combination. By default all packages for the latest version are returned. If the v query string parameter is included the packages for the specified version are returned.

Example Request:
 
### /\<CHANNEL>/\<PRODUCT>/metadata
Get details for a particular package. The ACCEPT HTTP header with a value of application/json must be provided in the request for a JSON response to be returned...otherwise the response will be plain text.

This endpoint supports the following query string parameters:

p is the platform. Valid values are returned from the /platforms endpoint.

pv is the platform version. Possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15.

m is the machine architecture for the machine on which the product will be installed. Valid values are returned by the /architectures endpoint.

v is the version of the product to be installed. A version always takes the form x.y.z. Default value: latest.
 
### /\<CHANNEL>/\<PRODUCT>/download
Returns a 302 redirect to the download URL for a specific package. The following parameters must be provided. This is a perfect URL to use for the actual download buttons. This endpoint supports the same query string parameters as /<CHANNEL>/<PRODUCT>/metadata. Example Request:
 
## What is Download API - Opensource ?

When a user is having a free license they can connect to the `opensource` instance of the API and use that to provide them with product information and also be able to download the opensource version of the chef products for the which the license applies.

Valid `<CHANNEL>` values in endpoint URLs is limited to `stable`.

Endpoint results are restricted to opensource versions of products.



 
 
