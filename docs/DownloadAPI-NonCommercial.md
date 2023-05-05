# Download API Documentation (Non-Commercial License)
 There are two environments available for Download API i.e: `production`, `acceptance`. Below are the base url for emvironments. `Acceptance` environment is an internal testing environment for Download API
 ### Base url for Acceptance environments.
 
 `trial` : https://trial-acceptance.downloads.chef.co/
 
 `opensource` : https://opensource-acceptance.downloads.chef.co/

 ### Base url for Production environments.
 
  `trial` : https://ChefDownload-Trial.chef.io
 
 `opensource` : https://ChefDownload-Community.chef.io
 
 
 ## API Operation Modes

Download API has three different use cases : `trial`, `opensource`, and `commercial`. These use cases indicate the type of license which the user may have - 
`opensource`, `trial` , `commercial`. A user may be able to get a `commercial` license or `trial` or `free` license which will qualify them to use the 
product based on the license policies.

 ### Trial Mode

Valid `<CHANNEL>` values in endpoint URLs is limited to `stable` 
Endpoint results are limited to the most recent version of any product unless a valid commercial `license_id` is provided.

### Opensource Mode

Valid `<CHANNEL>` values in endpoint URLs is limited to `stable`.

Endpoint results are restricted to opensource versions of products. 

 
 
