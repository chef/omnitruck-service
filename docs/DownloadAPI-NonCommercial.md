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
product based on the license policies. The Download API for commercial licenses is discussed in the [link here{https://github.com/chef/omnitruck-service/blob/main/docs/DownloadAPI-Commercial.md}]

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
```
curl -X 'GET' \
  'https://chefdownload-trial.chef.io/stable/chef/versions/all?license_id=d8ed0e36-5d27-44b1-994b-e65f45c0704a&eol=false' \
  -H 'accept: application/json'
  
  [
  "16.0.257",
  "16.0.275",
  "16.0.287",
  "16.1.0",
  "16.1.16",
  "16.2.44",
  "16.2.50",
  "16.2.73",
  "16.3.38",
  "16.3.45",
  "16.4.35",
  "16.4.38",
  "16.4.41",
  "16.5.64",
  "16.5.77",
  "16.6.14",
  "16.7.61",
  "16.8.9",
  "16.8.14",
  "16.9.16",
  "16.9.17",
  "16.9.20",
  "16.9.29",
  "16.9.32",
  "16.10.8",
  "16.10.17",
  "16.11.7",
  "16.12.3",
  "16.13.16",
  "16.14.1",
  "16.15.22",
  "16.16.7",
  "16.16.13",
  "16.17.4",
  "16.17.18",
  "16.17.39",
  "16.17.51",
  "16.18.0",
  "16.18.30",
  "17.0.242",
  "17.1.35",
  "17.2.29",
  "17.3.48",
  "17.4.25",
  "17.4.38",
  "17.5.22",
  "17.6.15",
  "17.6.18",
  "17.7.22",
  "17.7.29",
  "17.8.25",
  "17.9.18",
  "17.9.26",
  "17.9.42",
  "17.9.46",
  "17.9.52",
  "17.10.0",
  "17.10.3",
  "18.0.185",
  "18.1.0",
  "18.2.7"
]

curl -X 'GET' \
  'https://chefdownload-trial.chef.io/stable/chef/versions/all?license_id=d8ed0e36-5d27-44b1-994b-e65f45c0704a&eol=true' \
  -H 'accept: application/json'
  
  [
  "10.12.0",
  "10.14.0",
  "10.14.2",
  "10.14.4",
  "10.16.0",
  "10.16.2",
  "10.16.4",
  "10.16.6",
  "10.18.0",
  "10.18.2",
  "10.20.0",
  "10.22.0",
  "10.24.0",
  "10.24.2",
  "10.24.4",
  "10.26.0",
  "10.28.0",
  "10.28.2",
  "10.30.2",
  "10.30.4",
  "10.32.2",
  "10.34.0",
  "10.34.2",
  "10.34.4",
  "10.34.6",
  "11.0.0",
  "11.2.0",
  "11.4.0",
  "11.4.2",
  "11.4.4",
  "11.6.0",
  "11.6.2",
  "11.8.0",
  "11.8.2",
  "11.10.0",
  "11.10.2",
  "11.10.4",
  "11.12.0",
  "11.12.2",
  "11.12.4",
  "11.12.8",
  "11.14.2",
  "11.14.6",
  "11.16.0",
  "11.16.2",
  "11.16.4",
  "11.18.0",
  "11.18.6",
  "11.18.10",
  "11.18.12",
  "12.0.0",
  "12.0.1",
  "12.0.3",
  "12.1.0",
  "12.1.1",
  "12.1.2",
  "12.2.0-rc.1",
  "12.2.1",
  "12.3.0-rc.0",
  "12.3.0",
  "12.4.0",
  "12.4.1",
  "12.4.2",
  "12.4.3",
  "12.5.1",
  "12.6.0",
  "12.7.2",
  "12.8.1",
  "12.9.38",
  "12.9.41",
  "12.10.24",
  "12.11.18",
  "12.12.13",
  "12.12.15",
  "12.13.30",
  "12.13.37",
  "12.14.60",
  "12.14.77",
  "12.14.89",
  "12.15.19",
  "12.16.42",
  "12.17.44",
  "12.18.31",
  "12.19.33",
  "12.19.36",
  "12.20.3",
  "12.21.1",
  "12.21.3",
  "12.21.4",
  "12.21.10",
  "12.21.12",
  "12.21.14",
  "12.21.20",
  "12.21.26",
  "12.21.31",
  "12.22.1",
  "12.22.3",
  "12.22.5",
  "13.0.113",
  "13.0.118",
  "13.1.31",
  "13.2.20",
  "13.3.42",
  "13.4.19",
  "13.4.24",
  "13.5.3",
  "13.6.0",
  "13.6.4",
  "13.7.16",
  "13.8.0",
  "13.8.3",
  "13.8.5",
  "13.9.1",
  "13.9.4",
  "13.10.0",
  "13.10.4",
  "13.11.3",
  "13.12.3",
  "13.12.14",
  "14.0.190",
  "14.0.202",
  "14.1.1",
  "14.1.12",
  "14.2.0",
  "14.3.37",
  "14.4.56",
  "14.5.27",
  "14.5.33",
  "14.6.47",
  "14.7.17",
  "14.8.12",
  "14.9.13",
  "14.10.9",
  "14.11.21",
  "14.12.3",
  "14.12.9",
  "14.13.11",
  "14.14.14",
  "14.14.25",
  "14.14.29",
  "14.15.6",
  "15.0.293",
  "15.0.298",
  "15.0.300",
  "15.1.36",
  "15.2.20",
  "15.3.14",
  "15.4.45",
  "15.5.9",
  "15.5.15",
  "15.5.16",
  "15.5.17",
  "15.6.10",
  "15.7.30",
  "15.7.31",
  "15.7.32",
  "15.8.23",
  "15.9.17",
  "15.10.12",
  "15.11.3",
  "15.11.8",
  "15.12.22",
  "15.13.8",
  "15.14.0",
  "15.15.0",
  "15.15.1",
  "15.16.2",
  "15.16.4",
  "15.16.7",
  "15.17.4",
  "16.0.257",
  "16.0.275",
  "16.0.287",
  "16.1.0",
  "16.1.16",
  "16.2.44",
  "16.2.50",
  "16.2.73",
  "16.3.38",
  "16.3.45",
  "16.4.35",
  "16.4.38",
  "16.4.41",
  "16.5.64",
  "16.5.77",
  "16.6.14",
  "16.7.61",
  "16.8.9",
  "16.8.14",
  "16.9.16",
  "16.9.17",
  "16.9.20",
  "16.9.29",
  "16.9.32",
  "16.10.8",
  "16.10.17",
  "16.11.7",
  "16.12.3",
  "16.13.16",
  "16.14.1",
  "16.15.22",
  "16.16.7",
  "16.16.13",
  "16.17.4",
  "16.17.18",
  "16.17.39",
  "16.17.51",
  "16.18.0",
  "16.18.30",
  "17.0.242",
  "17.1.35",
  "17.2.29",
  "17.3.48",
  "17.4.25",
  "17.4.38",
  "17.5.22",
  "17.6.15",
  "17.6.18",
  "17.7.22",
  "17.7.29",
  "17.8.25",
  "17.9.18",
  "17.9.26",
  "17.9.42",
  "17.9.46",
  "17.9.52",
  "17.10.0",
  "17.10.3",
  "18.0.185",
  "18.1.0",
  "18.2.7"
]
```

### /\<CHANNEL>/\<PRODUCT>/versions/latest
Get the latest version number for a particular channel and product combination.
Example Request:
```
curl -X 'GET' \
  'https://chefdownload-trial.chef.io/stable/chef/versions/latest?license_id=d8ed0e36-5d27-44b1-994b-e65f45c0704a' \
  -H 'accept: application/json'
  
  "18.2.7"
```
 
### /\<CHANNEL>/\<PRODUCT>/packages
Get the full list of all packages for a particular channel and product combination. By default all packages for the latest version are returned. If the v query string parameter is included the packages for the specified version are returned.

Example Request:
```
curl -X 'GET' \
  'https://chefdownload-trial.chef.io/stable/chef/packages?v=18.2.7&license_id=d8ed0e36-5d27-44b1-994b-e65f45c0704a&eol=false' \
  -H 'accept: application/json'
  {
  "amazon": {
    "2": {
      "aarch64": {
        "sha1": "66e215d0461d68ce1909deecb65096acfcb5226c",
        "sha256": "6def59a714fd6fb260967316868c8bc2729500dfa095f2e3e8197ed3c287bf56",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/amazon/2/chef-18.2.7-1.el7.aarch64.rpm",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "eddef044114a1d629b1d7886a89d4c9d222105ec",
        "sha256": "5a52c955db20f017a213838e6fb45af029c0e67e7e28d5fd7aca23cbec24d543",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/amazon/2/chef-18.2.7-1.el7.x86_64.rpm",
        "version": "18.2.7"
      }
    }
  },
  "debian": {
    "9": {
      "x86_64": {
        "sha1": "7462bffd901d85f4d9f6c54c15084a47732959d8",
        "sha256": "a4461840de71f08f11f3c65a6d2f40f41d394e98f84979f7a8388ed0b578c666",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/debian/9/chef_18.2.7-1_amd64.deb",
        "version": "18.2.7"
      }
    },
    "10": {
      "x86_64": {
        "sha1": "7462bffd901d85f4d9f6c54c15084a47732959d8",
        "sha256": "a4461840de71f08f11f3c65a6d2f40f41d394e98f84979f7a8388ed0b578c666",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/debian/10/chef_18.2.7-1_amd64.deb",
        "version": "18.2.7"
      }
    },
    "11": {
      "x86_64": {
        "sha1": "7462bffd901d85f4d9f6c54c15084a47732959d8",
        "sha256": "a4461840de71f08f11f3c65a6d2f40f41d394e98f84979f7a8388ed0b578c666",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/debian/11/chef_18.2.7-1_amd64.deb",
        "version": "18.2.7"
      }
    }
  },
  "el": {
    "6": {
      "x86_64": {
        "sha1": "517b3418aa70ec2418b36583ab21b5e3cefab027",
        "sha256": "1fd570690b2629fdff2d2771794a966c2b3aeed3321c3ff2af49454fe5baf792",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/6/chef-18.2.7-1.el6.x86_64.rpm",
        "version": "18.2.7"
      }
    },
    "7": {
      "aarch64": {
        "sha1": "66e215d0461d68ce1909deecb65096acfcb5226c",
        "sha256": "6def59a714fd6fb260967316868c8bc2729500dfa095f2e3e8197ed3c287bf56",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/7/chef-18.2.7-1.el7.aarch64.rpm",
        "version": "18.2.7"
      },
      "ppc64": {
        "sha1": "a02fbb6736bdd043a6b3a44a8871b684627f7db0",
        "sha256": "3882e5b4a431594ba70b7deb35b516c5f312221f722243a26fdfd576209c3450",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/7/chef-18.2.7-1.el7.ppc64.rpm",
        "version": "18.2.7"
      },
      "ppc64le": {
        "sha1": "9262de62b371db523a5345a70e322b15f8794521",
        "sha256": "b3c91b91d591de1580651fa0ddd71113b2326d0269f3177fa3728292151d8a15",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/7/chef-18.2.7-1.el7.ppc64le.rpm",
        "version": "18.2.7"
      },
      "s390x": {
        "sha1": "f26306f8c4a4990acfa7d7b85df7921c91ef014c",
        "sha256": "5ceda6f58c41b8c4de6b8be40e682a8984bb93f2efe07eae8bc8b1f25afb33ad",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/7/chef-18.2.7-1.el7.s390x.rpm",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "eddef044114a1d629b1d7886a89d4c9d222105ec",
        "sha256": "5a52c955db20f017a213838e6fb45af029c0e67e7e28d5fd7aca23cbec24d543",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/7/chef-18.2.7-1.el7.x86_64.rpm",
        "version": "18.2.7"
      }
    },
    "8": {
      "aarch64": {
        "sha1": "62cdb9a34eec9e851e3371adf3bdcccbfe17c552",
        "sha256": "1ef5a804fff72cc2332a475c6cb493d042722874098fb8a365615ac5626627e1",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/8/chef-18.2.7-1.el8.aarch64.rpm",
        "version": "18.2.7"
      },
      "s390x": {
        "sha1": "f26306f8c4a4990acfa7d7b85df7921c91ef014c",
        "sha256": "5ceda6f58c41b8c4de6b8be40e682a8984bb93f2efe07eae8bc8b1f25afb33ad",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/8/chef-18.2.7-1.el7.s390x.rpm",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "f8b31f3eb8d4153c3ed163aa88bdefc52acbb7d7",
        "sha256": "3991841f8e2f43b5ae4179e998149fbf97ab33c9af1e53b7f1f97638bf797271",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/8/chef-18.2.7-1.el8.x86_64.rpm",
        "version": "18.2.7"
      }
    },
    "9": {
      "aarch64": {
        "sha1": "1e8ce841b63714099e003c6a1b60854f3857d0ce",
        "sha256": "5933c7f8e98716e26b62ab67fbf3eaa7a4c7df864bd1b1d23372cfa2a3e233da",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/9/chef-18.2.7-1.el9.aarch64.rpm",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "c57d95d0733cd68b9fddc12255a15103abf2e1f6",
        "sha256": "5f71a6db0c8189d28b68b763261da99d89c45676eb557f39d6a2ed7456dbc09b",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/9/chef-18.2.7-1.el9.x86_64.rpm",
        "version": "18.2.7"
      }
    }
  },
  "freebsd": {
    "12": {
      "x86_64": {
        "sha1": "28de4be1ba5a0c72783f6584cf302a70eb6d1bab",
        "sha256": "328633a33a7c2f17541f7029bf93c5f8c3c394284cf49944930e60e101a35461",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/freebsd/12/chef-18.2.7_1.amd64.sh",
        "version": "18.2.7"
      }
    }
  },
  "mac_os_x": {
    "11": {
      "aarch64": {
        "sha1": "6b21006f632be9415bca96e4fe630ab43e50e070",
        "sha256": "9c32bfa3648548ac496ad592054a13bb3eb037dbf68ec4a594333407cf0df2b3",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/mac_os_x/11/chef-18.2.7-1.arm64.dmg",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "353e0a31a3a70c8cbf342affec654361ef593af6",
        "sha256": "7a49c3a92f808daf1ea53ed6077feb1b9c371fa18b4ee6e6032aded79741fecc",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/mac_os_x/11/chef-18.2.7-1.x86_64.dmg",
        "version": "18.2.7"
      }
    },
    "12": {
      "aarch64": {
        "sha1": "6b21006f632be9415bca96e4fe630ab43e50e070",
        "sha256": "9c32bfa3648548ac496ad592054a13bb3eb037dbf68ec4a594333407cf0df2b3",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/mac_os_x/12/chef-18.2.7-1.arm64.dmg",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "353e0a31a3a70c8cbf342affec654361ef593af6",
        "sha256": "7a49c3a92f808daf1ea53ed6077feb1b9c371fa18b4ee6e6032aded79741fecc",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/mac_os_x/12/chef-18.2.7-1.x86_64.dmg",
        "version": "18.2.7"
      }
    },
    "10.15": {
      "x86_64": {
        "sha1": "353e0a31a3a70c8cbf342affec654361ef593af6",
        "sha256": "7a49c3a92f808daf1ea53ed6077feb1b9c371fa18b4ee6e6032aded79741fecc",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/mac_os_x/10.15/chef-18.2.7-1.x86_64.dmg",
        "version": "18.2.7"
      }
    }
  },
  "sles": {
    "12": {
      "s390x": {
        "sha1": "e375c1d4e839af4decc0c10b1a36158ef0d9104e",
        "sha256": "e87658844212187d14c092933f4414f2a5cb1f6d0d7996657a21fb0d3c813eba",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/sles/12/chef-18.2.7-1.sles12.s390x.rpm",
        "version": "18.2.7"
      }
    },
    "15": {
      "aarch64": {
        "sha1": "673abd5f929ccf48c2c0ce74b5c9eff17e5e973f",
        "sha256": "a0c9b35f1a9bca8490ecd101ddacc51de5d3eaf3ece63c1160e323c3c0b6a3b0",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/sles/15/chef-18.2.7-1.sles15.aarch64.rpm",
        "version": "18.2.7"
      },
      "s390x": {
        "sha1": "e375c1d4e839af4decc0c10b1a36158ef0d9104e",
        "sha256": "e87658844212187d14c092933f4414f2a5cb1f6d0d7996657a21fb0d3c813eba",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/sles/15/chef-18.2.7-1.sles12.s390x.rpm",
        "version": "18.2.7"
      }
    }
  },
  "solaris2": {
    "5.11": {
      "i386": {
        "sha1": "eae24daf9632ee314f18de3a1f99cf3fe2e41094",
        "sha256": "7d441e876a66a8623eb90dbb6b5d3e3b17445db26ab60844f7b53b4743b7e7d4",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/solaris2/5.11/chef-18.2.7-1.i386.p5p",
        "version": "18.2.7"
      },
      "sparc": {
        "sha1": "14e1d22d5838dfc51753bf0ac7f03cb90cabda82",
        "sha256": "0e765655fd810707f08c1d5e279080739e3d4eea9ce0566b8c71b1e74093f91a",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/solaris2/5.11/chef-18.2.7-1.sparc.p5p",
        "version": "18.2.7"
      }
    }
  },
  "ubuntu": {
    "16.04": {
      "x86_64": {
        "sha1": "8e8ae315d4695f9c95efc0a1437d2d453f7ab116",
        "sha256": "86f14ae08237b4e24201436ecb83c08c29b68aed1d6ede0953a1b4547a920e36",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/ubuntu/16.04/chef_18.2.7-1_amd64.deb",
        "version": "18.2.7"
      }
    },
    "18.04": {
      "aarch64": {
        "sha1": "0bca58ac38a1818eb0f86079d1c4a8158687b852",
        "sha256": "684a25f537fcc3cab0a10b7345c3484a7dd025dec07668f9785b8ec6db01db61",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/ubuntu/18.04/chef_18.2.7-1_arm64.deb",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "8e8ae315d4695f9c95efc0a1437d2d453f7ab116",
        "sha256": "86f14ae08237b4e24201436ecb83c08c29b68aed1d6ede0953a1b4547a920e36",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/ubuntu/18.04/chef_18.2.7-1_amd64.deb",
        "version": "18.2.7"
      }
    },
    "20.04": {
      "aarch64": {
        "sha1": "0bca58ac38a1818eb0f86079d1c4a8158687b852",
        "sha256": "684a25f537fcc3cab0a10b7345c3484a7dd025dec07668f9785b8ec6db01db61",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/ubuntu/20.04/chef_18.2.7-1_arm64.deb",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "8e8ae315d4695f9c95efc0a1437d2d453f7ab116",
        "sha256": "86f14ae08237b4e24201436ecb83c08c29b68aed1d6ede0953a1b4547a920e36",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/ubuntu/20.04/chef_18.2.7-1_amd64.deb",
        "version": "18.2.7"
      }
    },
    "22.04": {
      "aarch64": {
        "sha1": "0bca58ac38a1818eb0f86079d1c4a8158687b852",
        "sha256": "684a25f537fcc3cab0a10b7345c3484a7dd025dec07668f9785b8ec6db01db61",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/ubuntu/22.04/chef_18.2.7-1_arm64.deb",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "8e8ae315d4695f9c95efc0a1437d2d453f7ab116",
        "sha256": "86f14ae08237b4e24201436ecb83c08c29b68aed1d6ede0953a1b4547a920e36",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/ubuntu/22.04/chef_18.2.7-1_amd64.deb",
        "version": "18.2.7"
      }
    }
  },
  "windows": {
    "10": {
      "x86_64": {
        "sha1": "9fe79bbdbad6d8d3e33fa84f3f97be7a834a6f86",
        "sha256": "6aadc330f31093871c9a5d8ef09c3d64bbb867b9e1c8eb6b7f3601e5f888b323",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/windows/10/chef-client-18.2.7-1-x64.msi",
        "version": "18.2.7"
      }
    },
    "11": {
      "x86_64": {
        "sha1": "9fe79bbdbad6d8d3e33fa84f3f97be7a834a6f86",
        "sha256": "6aadc330f31093871c9a5d8ef09c3d64bbb867b9e1c8eb6b7f3601e5f888b323",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/windows/11/chef-client-18.2.7-1-x64.msi",
        "version": "18.2.7"
      }
    },
    "2012": {
      "x86_64": {
        "sha1": "9fe79bbdbad6d8d3e33fa84f3f97be7a834a6f86",
        "sha256": "6aadc330f31093871c9a5d8ef09c3d64bbb867b9e1c8eb6b7f3601e5f888b323",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/windows/2012/chef-client-18.2.7-1-x64.msi",
        "version": "18.2.7"
      }
    },
    "2016": {
      "x86_64": {
        "sha1": "9fe79bbdbad6d8d3e33fa84f3f97be7a834a6f86",
        "sha256": "6aadc330f31093871c9a5d8ef09c3d64bbb867b9e1c8eb6b7f3601e5f888b323",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/windows/2016/chef-client-18.2.7-1-x64.msi",
        "version": "18.2.7"
      }
    },
    "2019": {
      "x86_64": {
        "sha1": "9fe79bbdbad6d8d3e33fa84f3f97be7a834a6f86",
        "sha256": "6aadc330f31093871c9a5d8ef09c3d64bbb867b9e1c8eb6b7f3601e5f888b323",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/windows/2019/chef-client-18.2.7-1-x64.msi",
        "version": "18.2.7"
      }
    },
    "2022": {
      "x86_64": {
        "sha1": "9fe79bbdbad6d8d3e33fa84f3f97be7a834a6f86",
        "sha256": "6aadc330f31093871c9a5d8ef09c3d64bbb867b9e1c8eb6b7f3601e5f888b323",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/windows/2022/chef-client-18.2.7-1-x64.msi",
        "version": "18.2.7"
      }
    },
    "2012r2": {
      "x86_64": {
        "sha1": "9fe79bbdbad6d8d3e33fa84f3f97be7a834a6f86",
        "sha256": "6aadc330f31093871c9a5d8ef09c3d64bbb867b9e1c8eb6b7f3601e5f888b323",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/windows/2012r2/chef-client-18.2.7-1-x64.msi",
        "version": "18.2.7"
      }
    }
  }
}

curl -X 'GET' \
  'https://chefdownload-trial.chef.io/stable/chef/packages?v=18.2.7&license_id=d8ed0e36-5d27-44b1-994b-e65f45c0704a&eol=true' \
  -H 'accept: application/json'
  {
  "amazon": {
    "2": {
      "aarch64": {
        "sha1": "66e215d0461d68ce1909deecb65096acfcb5226c",
        "sha256": "6def59a714fd6fb260967316868c8bc2729500dfa095f2e3e8197ed3c287bf56",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/amazon/2/chef-18.2.7-1.el7.aarch64.rpm",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "eddef044114a1d629b1d7886a89d4c9d222105ec",
        "sha256": "5a52c955db20f017a213838e6fb45af029c0e67e7e28d5fd7aca23cbec24d543",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/amazon/2/chef-18.2.7-1.el7.x86_64.rpm",
        "version": "18.2.7"
      }
    }
  },
  "debian": {
    "9": {
      "x86_64": {
        "sha1": "7462bffd901d85f4d9f6c54c15084a47732959d8",
        "sha256": "a4461840de71f08f11f3c65a6d2f40f41d394e98f84979f7a8388ed0b578c666",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/debian/9/chef_18.2.7-1_amd64.deb",
        "version": "18.2.7"
      }
    },
    "10": {
      "x86_64": {
        "sha1": "7462bffd901d85f4d9f6c54c15084a47732959d8",
        "sha256": "a4461840de71f08f11f3c65a6d2f40f41d394e98f84979f7a8388ed0b578c666",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/debian/10/chef_18.2.7-1_amd64.deb",
        "version": "18.2.7"
      }
    },
    "11": {
      "x86_64": {
        "sha1": "7462bffd901d85f4d9f6c54c15084a47732959d8",
        "sha256": "a4461840de71f08f11f3c65a6d2f40f41d394e98f84979f7a8388ed0b578c666",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/debian/11/chef_18.2.7-1_amd64.deb",
        "version": "18.2.7"
      }
    }
  },
  "el": {
    "6": {
      "x86_64": {
        "sha1": "517b3418aa70ec2418b36583ab21b5e3cefab027",
        "sha256": "1fd570690b2629fdff2d2771794a966c2b3aeed3321c3ff2af49454fe5baf792",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/6/chef-18.2.7-1.el6.x86_64.rpm",
        "version": "18.2.7"
      }
    },
    "7": {
      "aarch64": {
        "sha1": "66e215d0461d68ce1909deecb65096acfcb5226c",
        "sha256": "6def59a714fd6fb260967316868c8bc2729500dfa095f2e3e8197ed3c287bf56",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/7/chef-18.2.7-1.el7.aarch64.rpm",
        "version": "18.2.7"
      },
      "ppc64": {
        "sha1": "a02fbb6736bdd043a6b3a44a8871b684627f7db0",
        "sha256": "3882e5b4a431594ba70b7deb35b516c5f312221f722243a26fdfd576209c3450",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/7/chef-18.2.7-1.el7.ppc64.rpm",
        "version": "18.2.7"
      },
      "ppc64le": {
        "sha1": "9262de62b371db523a5345a70e322b15f8794521",
        "sha256": "b3c91b91d591de1580651fa0ddd71113b2326d0269f3177fa3728292151d8a15",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/7/chef-18.2.7-1.el7.ppc64le.rpm",
        "version": "18.2.7"
      },
      "s390x": {
        "sha1": "f26306f8c4a4990acfa7d7b85df7921c91ef014c",
        "sha256": "5ceda6f58c41b8c4de6b8be40e682a8984bb93f2efe07eae8bc8b1f25afb33ad",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/7/chef-18.2.7-1.el7.s390x.rpm",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "eddef044114a1d629b1d7886a89d4c9d222105ec",
        "sha256": "5a52c955db20f017a213838e6fb45af029c0e67e7e28d5fd7aca23cbec24d543",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/7/chef-18.2.7-1.el7.x86_64.rpm",
        "version": "18.2.7"
      }
    },
    "8": {
      "aarch64": {
        "sha1": "62cdb9a34eec9e851e3371adf3bdcccbfe17c552",
        "sha256": "1ef5a804fff72cc2332a475c6cb493d042722874098fb8a365615ac5626627e1",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/8/chef-18.2.7-1.el8.aarch64.rpm",
        "version": "18.2.7"
      },
      "s390x": {
        "sha1": "f26306f8c4a4990acfa7d7b85df7921c91ef014c",
        "sha256": "5ceda6f58c41b8c4de6b8be40e682a8984bb93f2efe07eae8bc8b1f25afb33ad",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/8/chef-18.2.7-1.el7.s390x.rpm",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "f8b31f3eb8d4153c3ed163aa88bdefc52acbb7d7",
        "sha256": "3991841f8e2f43b5ae4179e998149fbf97ab33c9af1e53b7f1f97638bf797271",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/8/chef-18.2.7-1.el8.x86_64.rpm",
        "version": "18.2.7"
      }
    },
    "9": {
      "aarch64": {
        "sha1": "1e8ce841b63714099e003c6a1b60854f3857d0ce",
        "sha256": "5933c7f8e98716e26b62ab67fbf3eaa7a4c7df864bd1b1d23372cfa2a3e233da",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/9/chef-18.2.7-1.el9.aarch64.rpm",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "c57d95d0733cd68b9fddc12255a15103abf2e1f6",
        "sha256": "5f71a6db0c8189d28b68b763261da99d89c45676eb557f39d6a2ed7456dbc09b",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/el/9/chef-18.2.7-1.el9.x86_64.rpm",
        "version": "18.2.7"
      }
    }
  },
  "freebsd": {
    "12": {
      "x86_64": {
        "sha1": "28de4be1ba5a0c72783f6584cf302a70eb6d1bab",
        "sha256": "328633a33a7c2f17541f7029bf93c5f8c3c394284cf49944930e60e101a35461",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/freebsd/12/chef-18.2.7_1.amd64.sh",
        "version": "18.2.7"
      }
    }
  },
  "mac_os_x": {
    "11": {
      "aarch64": {
        "sha1": "6b21006f632be9415bca96e4fe630ab43e50e070",
        "sha256": "9c32bfa3648548ac496ad592054a13bb3eb037dbf68ec4a594333407cf0df2b3",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/mac_os_x/11/chef-18.2.7-1.arm64.dmg",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "353e0a31a3a70c8cbf342affec654361ef593af6",
        "sha256": "7a49c3a92f808daf1ea53ed6077feb1b9c371fa18b4ee6e6032aded79741fecc",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/mac_os_x/11/chef-18.2.7-1.x86_64.dmg",
        "version": "18.2.7"
      }
    },
    "12": {
      "aarch64": {
        "sha1": "6b21006f632be9415bca96e4fe630ab43e50e070",
        "sha256": "9c32bfa3648548ac496ad592054a13bb3eb037dbf68ec4a594333407cf0df2b3",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/mac_os_x/12/chef-18.2.7-1.arm64.dmg",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "353e0a31a3a70c8cbf342affec654361ef593af6",
        "sha256": "7a49c3a92f808daf1ea53ed6077feb1b9c371fa18b4ee6e6032aded79741fecc",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/mac_os_x/12/chef-18.2.7-1.x86_64.dmg",
        "version": "18.2.7"
      }
    },
    "10.15": {
      "x86_64": {
        "sha1": "353e0a31a3a70c8cbf342affec654361ef593af6",
        "sha256": "7a49c3a92f808daf1ea53ed6077feb1b9c371fa18b4ee6e6032aded79741fecc",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/mac_os_x/10.15/chef-18.2.7-1.x86_64.dmg",
        "version": "18.2.7"
      }
    }
  },
  "sles": {
    "12": {
      "s390x": {
        "sha1": "e375c1d4e839af4decc0c10b1a36158ef0d9104e",
        "sha256": "e87658844212187d14c092933f4414f2a5cb1f6d0d7996657a21fb0d3c813eba",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/sles/12/chef-18.2.7-1.sles12.s390x.rpm",
        "version": "18.2.7"
      }
    },
    "15": {
      "aarch64": {
        "sha1": "673abd5f929ccf48c2c0ce74b5c9eff17e5e973f",
        "sha256": "a0c9b35f1a9bca8490ecd101ddacc51de5d3eaf3ece63c1160e323c3c0b6a3b0",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/sles/15/chef-18.2.7-1.sles15.aarch64.rpm",
        "version": "18.2.7"
      },
      "s390x": {
        "sha1": "e375c1d4e839af4decc0c10b1a36158ef0d9104e",
        "sha256": "e87658844212187d14c092933f4414f2a5cb1f6d0d7996657a21fb0d3c813eba",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/sles/15/chef-18.2.7-1.sles12.s390x.rpm",
        "version": "18.2.7"
      }
    }
  },
  "solaris2": {
    "5.11": {
      "i386": {
        "sha1": "eae24daf9632ee314f18de3a1f99cf3fe2e41094",
        "sha256": "7d441e876a66a8623eb90dbb6b5d3e3b17445db26ab60844f7b53b4743b7e7d4",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/solaris2/5.11/chef-18.2.7-1.i386.p5p",
        "version": "18.2.7"
      },
      "sparc": {
        "sha1": "14e1d22d5838dfc51753bf0ac7f03cb90cabda82",
        "sha256": "0e765655fd810707f08c1d5e279080739e3d4eea9ce0566b8c71b1e74093f91a",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/solaris2/5.11/chef-18.2.7-1.sparc.p5p",
        "version": "18.2.7"
      }
    }
  },
  "ubuntu": {
    "16.04": {
      "x86_64": {
        "sha1": "8e8ae315d4695f9c95efc0a1437d2d453f7ab116",
        "sha256": "86f14ae08237b4e24201436ecb83c08c29b68aed1d6ede0953a1b4547a920e36",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/ubuntu/16.04/chef_18.2.7-1_amd64.deb",
        "version": "18.2.7"
      }
    },
    "18.04": {
      "aarch64": {
        "sha1": "0bca58ac38a1818eb0f86079d1c4a8158687b852",
        "sha256": "684a25f537fcc3cab0a10b7345c3484a7dd025dec07668f9785b8ec6db01db61",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/ubuntu/18.04/chef_18.2.7-1_arm64.deb",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "8e8ae315d4695f9c95efc0a1437d2d453f7ab116",
        "sha256": "86f14ae08237b4e24201436ecb83c08c29b68aed1d6ede0953a1b4547a920e36",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/ubuntu/18.04/chef_18.2.7-1_amd64.deb",
        "version": "18.2.7"
      }
    },
    "20.04": {
      "aarch64": {
        "sha1": "0bca58ac38a1818eb0f86079d1c4a8158687b852",
        "sha256": "684a25f537fcc3cab0a10b7345c3484a7dd025dec07668f9785b8ec6db01db61",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/ubuntu/20.04/chef_18.2.7-1_arm64.deb",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "8e8ae315d4695f9c95efc0a1437d2d453f7ab116",
        "sha256": "86f14ae08237b4e24201436ecb83c08c29b68aed1d6ede0953a1b4547a920e36",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/ubuntu/20.04/chef_18.2.7-1_amd64.deb",
        "version": "18.2.7"
      }
    },
    "22.04": {
      "aarch64": {
        "sha1": "0bca58ac38a1818eb0f86079d1c4a8158687b852",
        "sha256": "684a25f537fcc3cab0a10b7345c3484a7dd025dec07668f9785b8ec6db01db61",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/ubuntu/22.04/chef_18.2.7-1_arm64.deb",
        "version": "18.2.7"
      },
      "x86_64": {
        "sha1": "8e8ae315d4695f9c95efc0a1437d2d453f7ab116",
        "sha256": "86f14ae08237b4e24201436ecb83c08c29b68aed1d6ede0953a1b4547a920e36",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/ubuntu/22.04/chef_18.2.7-1_amd64.deb",
        "version": "18.2.7"
      }
    }
  },
  "windows": {
    "10": {
      "x86_64": {
        "sha1": "9fe79bbdbad6d8d3e33fa84f3f97be7a834a6f86",
        "sha256": "6aadc330f31093871c9a5d8ef09c3d64bbb867b9e1c8eb6b7f3601e5f888b323",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/windows/10/chef-client-18.2.7-1-x64.msi",
        "version": "18.2.7"
      }
    },
    "11": {
      "x86_64": {
        "sha1": "9fe79bbdbad6d8d3e33fa84f3f97be7a834a6f86",
        "sha256": "6aadc330f31093871c9a5d8ef09c3d64bbb867b9e1c8eb6b7f3601e5f888b323",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/windows/11/chef-client-18.2.7-1-x64.msi",
        "version": "18.2.7"
      }
    },
    "2012": {
      "x86_64": {
        "sha1": "9fe79bbdbad6d8d3e33fa84f3f97be7a834a6f86",
        "sha256": "6aadc330f31093871c9a5d8ef09c3d64bbb867b9e1c8eb6b7f3601e5f888b323",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/windows/2012/chef-client-18.2.7-1-x64.msi",
        "version": "18.2.7"
      }
    },
    "2016": {
      "x86_64": {
        "sha1": "9fe79bbdbad6d8d3e33fa84f3f97be7a834a6f86",
        "sha256": "6aadc330f31093871c9a5d8ef09c3d64bbb867b9e1c8eb6b7f3601e5f888b323",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/windows/2016/chef-client-18.2.7-1-x64.msi",
        "version": "18.2.7"
      }
    },
    "2019": {
      "x86_64": {
        "sha1": "9fe79bbdbad6d8d3e33fa84f3f97be7a834a6f86",
        "sha256": "6aadc330f31093871c9a5d8ef09c3d64bbb867b9e1c8eb6b7f3601e5f888b323",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/windows/2019/chef-client-18.2.7-1-x64.msi",
        "version": "18.2.7"
      }
    },
    "2022": {
      "x86_64": {
        "sha1": "9fe79bbdbad6d8d3e33fa84f3f97be7a834a6f86",
        "sha256": "6aadc330f31093871c9a5d8ef09c3d64bbb867b9e1c8eb6b7f3601e5f888b323",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/windows/2022/chef-client-18.2.7-1-x64.msi",
        "version": "18.2.7"
      }
    },
    "2012r2": {
      "x86_64": {
        "sha1": "9fe79bbdbad6d8d3e33fa84f3f97be7a834a6f86",
        "sha256": "6aadc330f31093871c9a5d8ef09c3d64bbb867b9e1c8eb6b7f3601e5f888b323",
        "url": "https://packages.chef.io/files/stable/chef/18.2.7/windows/2012r2/chef-client-18.2.7-1-x64.msi",
        "version": "18.2.7"
      }
    }
  }
}

```
 
### /\<CHANNEL>/\<PRODUCT>/metadata
Get details for a particular package. The ACCEPT HTTP header with a value of application/json must be provided in the request for a JSON response to be returned...otherwise the response will be plain text.

This endpoint supports the following query string parameters:

p is the platform. Valid values are returned from the /platforms endpoint.

pv is the platform version. Possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15.

m is the machine architecture for the machine on which the product will be installed. Valid values are returned by the /architectures endpoint.

v is the version of the product to be installed. A version always takes the form x.y.z. Default value: latest.
Example Request:

```
curl -X 'GET' \
  'https://chefdownload-trial.chef.io/stable/chef/metadata?p=ubuntu&pv=18.04&m=x86_64&v=latest&license_id=d8ed0e36-5d27-44b1-994b-e65f45c0704a&eol=false' \
  -H 'accept: application/json'
  
  {
  "sha1": "8e8ae315d4695f9c95efc0a1437d2d453f7ab116",
  "sha256": "86f14ae08237b4e24201436ecb83c08c29b68aed1d6ede0953a1b4547a920e36",
  "url": "https://chefdownload-trial.chef.io/stable/chef/download?license_id=d8ed0e36-5d27-44b1-994b-e65f45c0704a&m=x86_64&p=ubuntu&pv=18.04&v=latest",
  "version": "18.2.7"
}

curl -X 'GET' \
  'https://chefdownload-trial.chef.io/stable/chef/metadata?p=ubuntu&pv=18.04&m=x86_64&v=latest&eol=true' \
  -H 'accept: application/json'
  
  {
  "sha1": "8e8ae315d4695f9c95efc0a1437d2d453f7ab116",
  "sha256": "86f14ae08237b4e24201436ecb83c08c29b68aed1d6ede0953a1b4547a920e36",
  "url": "https://chefdownload-trial.chef.io/stable/chef/download?license_id=&m=x86_64&p=ubuntu&pv=18.04&v=latest",
  "version": "18.2.7"
}

```

### /\<CHANNEL>/\<PRODUCT>/download
Returns a 302 redirect to the download URL for a specific package. The following parameters must be provided. This is a perfect URL to use for the actual download buttons. This endpoint supports the same query string parameters as /<CHANNEL>/<PRODUCT>/metadata. 
 
 Example Request:
 
 
## What is Download API - Opensource ?

When a user is having a free license they can connect to the `opensource` instance of the API and use that to provide them with product information and also be able to download the opensource version of the chef products for the which the license applies.

Valid `<CHANNEL>` values in endpoint URLs is limited to `stable`.

Endpoint results are restricted to opensource versions of products.



 
 
