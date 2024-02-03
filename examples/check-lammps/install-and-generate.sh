#!/bin/bash

# This assumes binding the entire directory with this script and lammps-experiment.yaml
hasGpu="${1:-no}"
path="${2:-./compatibility-spec.json}"
# TODO need arm/ppc64le here

# Compile compspec here so it's with the right glibc
# assume we have wget and git for now
# apt-get update && apt-get install -y wget git || yum update && yum install -y wget git
wget https://go.dev/dl/go1.20.3.linux-amd64.tar.gz
tar -xzf go1.20.3.linux-amd64.tar.gz  2>/dev/null
mv go /usr/local && rm go1.20.3.linux-amd64.tar.gz    
export PATH=$PATH:/usr/local/go/bin

git clone -b add-start-of-check-graph https://github.com/supercontainers/compspec-go /tmp/cs
cd /tmp/cs
make
export PATH=$PWD/bin:$PATH
cd -

# Download the spec for our compatibility artifact
wget --quiet https://gist.githubusercontent.com/vsoch/fcd0f7d633860674cb085a8540ce4bb2/raw/02290df3aa3439caf9754d118a612906be3e3594/lammps-experiment.yaml

# Generate!
compspec create --in ./lammps-experiment.yaml -a custom.gpu.available=$hasGpu -o ${path}
cat ${path}