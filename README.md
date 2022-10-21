# Licensed Omnitruck API

Omnitruck API service that provides license validation and entitlement checking for omnitruck requests.

## Requirements

* golang 1.19+

## Getting Started

Building the service and swagger documentation

```bash
make all
```

To just build the service without updating the swagger documentation

```bash
make build
```

Running the service

```bash
$ bin/omnitruck-service start
INFO[0000] Starting OpensourceServer                     pkg=cmd/opensource
```

## License

```
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
