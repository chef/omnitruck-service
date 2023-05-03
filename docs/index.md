# Download API Documentation
 There are two environments available for Download API i.e: `production`, `acceptance`. Below are the base url for emvironments. 'Acceptance' environment is an internal testing environment for Download API
 Base url for `Acceptance` environments.
 `commercial` : https://commercial-acceptance.downloads.chef.co/
 `trial` : https://trial-acceptance.downloads.chef.co/
 `opensource` : https://opensource-acceptance.downloads.chef.co/

 Base url for `Production` environments.
 `commercial` : https://ChefDownload-Commerical.chef.io
 `trial` : https://ChefDownload-Trial.chef.io
 `opensource` : https://ChefDownload-Community.chef.io

## API Operation Modes
---

Omnitruck API operates in 3 different modes; `trial`, `opensource`, and `commercial`. These modes indicate the type of license which the user may have - `free`, `trial` , `commercial`
### Trial Mode

Valid `<CHANNEL>` values in endpoint URLs is limited to `stable` 
Endpoint results are limited to the most recent version of any product unless a valid commercial `license_id` is provided.

### Opensource Mode

Valid `<CHANNEL>` values in endpoint URLs is limited to `stable`.

Endpoint results are restricted to opensource versions of products. 

### Commercial Mode

Valid `<CHANNEL>` values in endpoint URLs include: `current` and `stable`.

If `eol=true` url query parameter is returned than all product versions are returned, otherwise only supported products and product versions are returned.

Valid `license_id` url query parameter is required to return results from API endpoints.

## API Endpoints
---

### /products
Returns a valid list of valid product keys. Any of these product keys can be used in the \<PRODUCT> value of other endpoints. Please note many of these products are used for internal tools only and many have been EOL'd.

Example Request:

```bash
curl <BASE_URL>/products
[
  "analytics",
  "angry-omnibus-toolchain",
  "angrychef",
  "automate",
  "chef",
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


### /platforms
Returns a valid list of valid platform keys along with full friendly names. Any of these platform keys can be used in the p query string value in various endpoints below. 

Example Request:

```bash
curl <BASE_URL>/platforms
{
  "aix": "AIX",
  "amazon": "Amazon Linux",
  "el": "Red Hat Enterprise Linux/CentOS",
  "debian": "Debian GNU/Linux",
  "freebsd": "FreeBSD",
  "ios_xr": "Cisco IOS-XR",
  "mac_os_x": "macOS",
  "nexus": "Cisco NX-OS",
  "ubuntu": "Ubuntu Linux",
  "solaris2": "Solaris",
  "sles": "SUSE Linux Enterprise Server",
  "suse": "openSUSE",
  "windows": "Windows"
}
```

### /architectures
Returns a valid list of valid platform keys along with friendly names. Any of these platform keys can be used in the m query string value in various endpoints below. 

Example Request:

```bash
curl <BASE_URL>/architectures
[
  "aarch64",
  "armv7l",
  "i386",
  "powerpc",
  "ppc64",
  "ppc64le",
  "s390x",
  "sparc",
  "x86_64"
]
```

### /\<CHANNEL>/\<PRODUCT>/versions/all
Get a list of all available version numbers for a particular channel and product combination

Example Request:

```bash
curl <BASE_URL>/stable/chef-workstation/versions/all
[
  "0.1.119",
  "0.1.120",
  "0.1.133",
  "0.1.137",
  "0.1.139",
  "0.1.142",
  "0.1.148",
  "0.1.150",
  "0.1.155",
  "0.1.162",
  "0.2.21",
  "0.2.27",
  "0.2.29",
  "0.2.35",
  "0.2.39",
  "0.2.40",
  "0.2.41",
  "0.2.43",
  "0.2.48",
  "0.2.53",
  "0.3.2",
  "0.4.1",
  "0.4.2",
  "0.5.1",
  "0.6.2",
  "0.7.4",
  "0.8.7",
  "0.9.42",
  "0.10.41",
  "0.11.21",
  "0.12.20",
  "0.13.35",
  "0.14.16",
  "0.15.6",
  "0.15.18",
  "0.16.31",
  "0.16.32",
  "0.16.33",
  "0.17.5"
]
```

### /\<CHANNEL>/\<PRODUCT>/versions/latest
Get the latest version number for a particular channel and product combination.

Example Request:

```bash
curl <BASE_URL>/stable/chef/versions/latest
"15.8.23"
```

### /\<CHANNEL>/\<PRODUCT>/packages
Get the full list of all packages for a particular channel and product combination. By default all packages for the latest version are returned. If the v query string parameter is included the packages for the specified version are returned.

Example Request:

```bash
curl <BASE_URL>/stable/chef-workstation/packages
{
  "ubuntu": {
    "16.04": {
      "x86_64": {
        "sha1": "5d89b5c2b2c86980b18e6114066802d554188683",
        "sha256": "c447bdc246312dd8883cd61894da60c28405efb2904240eca4a3fd9c4122fb3a",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.16.33/ubuntu/16.04/chef-workstation_0.16.33-1_amd64.deb",
        "version": "0.16.33"
      }
    },
    "18.04": {
      "x86_64": {
        "sha1": "5d89b5c2b2c86980b18e6114066802d554188683",
        "sha256": "c447bdc246312dd8883cd61894da60c28405efb2904240eca4a3fd9c4122fb3a",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.16.33/ubuntu/18.04/chef-workstation_0.16.33-1_amd64.deb",
        "version": "0.16.33"
      }
    }
  },
  "el": {
    "6": {
      "x86_64": {
        "sha1": "36dc2460427bd4856e3a3a247747ebe25d8f545f",
        "sha256": "60377244bf500e5324a2c80ef3dc4a5e82fcae12d20626fdf22324b9e14ddecd",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.16.33/el/6/chef-workstation-0.16.33-1.el6.x86_64.rpm",
        "version": "0.16.33"
      }
    },
    "7": {
      "x86_64": {
        "sha1": "f6ac8f616624b41b9328149ab6a9fdd31a1466e5",
        "sha256": "a5e816c712fa90ac044c048d892f147274fafa5a6db070200a4945f6dc1e12b2",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.16.33/el/7/chef-workstation-0.16.33-1.el7.x86_64.rpm",
        "version": "0.16.33"
      }
    },
    "8": {
      "x86_64": {
        "sha1": "f6ac8f616624b41b9328149ab6a9fdd31a1466e5",
        "sha256": "a5e816c712fa90ac044c048d892f147274fafa5a6db070200a4945f6dc1e12b2",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.16.33/el/8/chef-workstation-0.16.33-1.el7.x86_64.rpm",
        "version": "0.16.33"
      }
    }
  },
  "mac_os_x": {
    "10.13": {
      "x86_64": {
        "sha1": "d87acebcf448cbbfbc20ce0863e963c93735627a",
        "sha256": "73ae4f3a6a22431d7d51f0a633121d1a2de65b4af9d3034ac7585cb9331711db",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.16.33/mac_os_x/10.13/chef-workstation-0.16.33-1.dmg",
        "version": "0.16.33"
      }
    },
    "10.14": {
      "x86_64": {
        "sha1": "d87acebcf448cbbfbc20ce0863e963c93735627a",
        "sha256": "73ae4f3a6a22431d7d51f0a633121d1a2de65b4af9d3034ac7585cb9331711db",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.16.33/mac_os_x/10.14/chef-workstation-0.16.33-1.dmg",
        "version": "0.16.33"
      }
    },
    "10.15": {
      "x86_64": {
        "sha1": "d87acebcf448cbbfbc20ce0863e963c93735627a",
        "sha256": "73ae4f3a6a22431d7d51f0a633121d1a2de65b4af9d3034ac7585cb9331711db",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.16.33/mac_os_x/10.15/chef-workstation-0.16.33-1.dmg",
        "version": "0.16.33"
      }
    }
  },
  "windows": {
    "2008r2": {
      "x86_64": {
        "sha1": "80ef10011ca6d170fe874612c9f3d64291824ad8",
        "sha256": "8556fab7e35cc14f67bc78867722b165adfd1403bdba7314cc894509c77afebd",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.16.33/windows/2019/chef-workstation-0.16.33-1-x64.msi",
        "version": "0.16.33"
      }
    }
  },
  "debian": {
    "9": {
      "x86_64": {
        "sha1": "29a13385be79d7ed6e3af1fae7417b3ce60e6a9f",
        "sha256": "3db05ca26e0b64b5c5a1c198d40036b297729e56b2f60e22243a06ab7699dfc0",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.16.33/debian/9/chef-workstation_0.16.33-1_amd64.deb",
        "version": "0.16.33"
      }
    },
    "8": {
      "x86_64": {
        "sha1": "29a13385be79d7ed6e3af1fae7417b3ce60e6a9f",
        "sha256": "3db05ca26e0b64b5c5a1c198d40036b297729e56b2f60e22243a06ab7699dfc0",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.16.33/debian/8/chef-workstation_0.16.33-1_amd64.deb",
        "version": "0.16.33"
      }
    },
    "10": {
      "x86_64": {
        "sha1": "29a13385be79d7ed6e3af1fae7417b3ce60e6a9f",
        "sha256": "3db05ca26e0b64b5c5a1c198d40036b297729e56b2f60e22243a06ab7699dfc0",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.16.33/debian/10/chef-workstation_0.16.33-1_amd64.deb",
        "version": "0.16.33"
      }
    }
  }
}
```

```bash
curl "<BASE_URL>/stable/chef-workstation/packages?v=0.1.133"
{
  "ubuntu": {
    "16.04": {
      "x86_64": {
        "sha1": "ce7b1dcc313e19669a92e820b2e1515985a96fdf",
        "sha256": "4a2171e5cebde6378b686ce2c7b35e8ff521b31eedfe7abb6b5179433592ded0",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.1.133/ubuntu/16.04/chef-workstation_0.1.133-1_amd64.deb",
        "version": "0.1.133"
      }
    },
    "14.04": {
      "x86_64": {
        "sha1": "ce7b1dcc313e19669a92e820b2e1515985a96fdf",
        "sha256": "4a2171e5cebde6378b686ce2c7b35e8ff521b31eedfe7abb6b5179433592ded0",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.1.133/ubuntu/14.04/chef-workstation_0.1.133-1_amd64.deb",
        "version": "0.1.133"
      }
    },
    "18.04": {
      "x86_64": {
        "sha1": "ce7b1dcc313e19669a92e820b2e1515985a96fdf",
        "sha256": "4a2171e5cebde6378b686ce2c7b35e8ff521b31eedfe7abb6b5179433592ded0",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.1.133/ubuntu/18.04/chef-workstation_0.1.133-1_amd64.deb",
        "version": "0.1.133"
      }
    }
  },
  "el": {
    "6": {
      "x86_64": {
        "sha1": "e122eef5264d52d18d8350565c8fb69e2fb0ddf6",
        "sha256": "0f82f830bccab46145dca39eef5326718586cc9f84f888fe26f0f3186e6bbaf2",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.1.133/el/6/chef-workstation-0.1.133-1.el6.x86_64.rpm",
        "version": "0.1.133"
      }
    },
    "7": {
      "x86_64": {
        "sha1": "e122eef5264d52d18d8350565c8fb69e2fb0ddf6",
        "sha256": "0f82f830bccab46145dca39eef5326718586cc9f84f888fe26f0f3186e6bbaf2",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.1.133/el/7/chef-workstation-0.1.133-1.el6.x86_64.rpm",
        "version": "0.1.133"
      }
    }
  },
  "mac_os_x": {
    "10.13": {
      "x86_64": {
        "sha1": "5c6c937b2beadc8d792834d9a6d937e0628a9abc",
        "sha256": "a676d7adfb0e8d668101f26793022b875a01eff696e093054f6bd118b6aa50d2",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.1.133/mac_os_x/10.13/chef-workstation-0.1.133-1.dmg",
        "version": "0.1.133"
      }
    },
    "10.11": {
      "x86_64": {
        "sha1": "5c6c937b2beadc8d792834d9a6d937e0628a9abc",
        "sha256": "a676d7adfb0e8d668101f26793022b875a01eff696e093054f6bd118b6aa50d2",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.1.133/mac_os_x/10.11/chef-workstation-0.1.133-1.dmg",
        "version": "0.1.133"
      }
    },
    "10.12": {
      "x86_64": {
        "sha1": "5c6c937b2beadc8d792834d9a6d937e0628a9abc",
        "sha256": "a676d7adfb0e8d668101f26793022b875a01eff696e093054f6bd118b6aa50d2",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.1.133/mac_os_x/10.12/chef-workstation-0.1.133-1.dmg",
        "version": "0.1.133"
      }
    }
  },
  "windows": {
    "2008r2": {
      "x86_64": {
        "sha1": "2cdb69fffef4546711b36692465ea2ba3cc1e054",
        "sha256": "cf956eb5e6e2ad9bce2cf022556e82a8e36f3568c6bfdd2b032a1a6d73350674",
        "url": "https://packages.chef.io/files/stable/chef-workstation/0.1.133/windows/2012r2/chef-workstation-0.1.133-1-x64.msi",
        "version": "0.1.133"
      }
    }
  }
}
```

### /\<CHANNEL>/\<PRODUCT>/metadata
 Get details for a particular package. The ACCEPT HTTP header with a value of application/json must be provided in the request for a JSON response to be returned...otherwise the response will be plain text. 

This endpoint supports the following query string parameters:

`p` is the platform. Valid values are returned from the /platforms endpoint.

`pv` is the platform version. Possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15.

`m` is the machine architecture for the machine on which the product will be installed. Valid values are returned by the /architectures endpoint.

`v` is the version of the product to be installed. A version always takes the form x.y.z. Default value: latest.

Example Request:

```bash
curl -H "ACCEPT:application/json"  "<BASE_URL>/stable/chef/metadata?p=mac_os_x&pv=10.15&m=x86_64"
{
  "sha1": "cac3e26b69ecca885cc54c61ff465954f5f148b9",
  "sha256": "f437427cd72ed14fb89590f4cdd4b252702bf5bf869c3c7450b323986d3cc6e3",
  "url": "https://packages.chef.io/files/stable/chef/15.8.23/mac_os_x/10.15/chef-15.8.23-1.dmg",
  "version": "15.8.23"
}
```

```bash
curl -H "ACCEPT:application/json" "<BASE_URL>/stable/chef/metadata?p=ubuntu&pv=18.04&m=x86_64"
{
  "sha1": "dc185e713e1dc3a79f699340c4fb169596375b43",
  "sha256": "d5a616db707690fe52aa90f52c13deb3e37c3b8790feb2c37154ab3c4565fda7",
  "url": "https://packages.chef.io/files/stable/chef/15.8.23/ubuntu/18.04/chef_15.8.23-1_amd64.deb",
  "version": "15.8.23"
}
```

### /\<CHANNEL>/\<PRODUCT>/download 
Returns a 302 redirect to the download URL for a specific package. The following parameters must be provided. This is a perfect URL to use for the actual download buttons. This endpoint supports the same query string parameters as /\<CHANNEL>/\<PRODUCT>/metadata.

Example Request:

```bash
curl -I "<BASE_URL>/stable/chef/download?p=mac_os_x&pv=10.15&m=x86_64"

HTTP/2 302
content-type: text/html;charset=utf-8
location: https://packages.chef.io/files/stable/chef/15.8.23/mac_os_x/10.15/chef-15.8.23-1.dmg
server: WEBrick/1.4.2 (Ruby/2.5.7/2019-10-01)
x-content-type-options: nosniff
x-frame-options: SAMEORIGIN
x-xss-protection: 1; mode=block
accept-ranges: bytes
date: Thu, 12 Mar 2020 19:57:08 GMT
via: 1.1 varnish
age: 37
x-served-by: cache-fty21372-FTY
x-cache: HIT
x-cache-hits: 1
x-timer: S1584043028.159446,VS0,VE0
content-length: 0
```
