#!/bin/bash
# Copyright (C) 2024, Nuklai. All rights reserved.
# See the file LICENSE for licensing terms.

yum update -y
yum install -y docker
service docker start
usermod -aG docker ec2-user
