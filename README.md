# apf

## Overview

This is a simple tool to fetch AWS prices from the AWS Price List API. It is intended to be command line tool.

## Installation

```bash
$ go get github.com/sfuruya0612/apf
```

## Usage

### Preparation

- AWS credentials
- Connection infomation to local MongoDB or MongoDB Atlas

### Fetch AWS Price

Assumes execution on local MongoDB

```bash
$ export MONGODB_URI="mongodb://localhost:27017"
$ apf fetch
```

### Get Price per service

#### Example

See help(`-h` option) for details

```bash
$ apf price --instance-type=t3.small ec2
Service   Region         OS/Engine InstanceType vCPU Memory PhysicalProcessor        ClockSpeed(GHz) Tenancy CapacityStatus PreInstalledSw ProcessorArchitecture OnDemandPrice(USD/hour) OnDemandPrice(USD/month)
AmazonEC2 ap-northeast-1 Linux     t3.small     2    2 GiB  Intel Skylake E5 2686 v5 3.1 GHz         Shared  Used           NA             64-bit                0.0272000000            19.86
```

```bash
$ apf price --vcpu=4 ec2
```

```bash
$ apf price --instance-type=t3.small ec2 --os=Windows
```
