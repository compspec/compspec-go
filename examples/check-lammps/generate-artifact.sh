#!/bin/bash

# This assumes binding the entire directory with this script and lammps-experiment.yaml
hasGpu="${1:-no}"
path="${2:-./compatibility-spec.json}"

# Note that this is hard coded for amd64, for arm you would wantt o add -arm or ppc64le -ppc
wget --quiet https://github.com/supercontainers/compspec-go/releases/download/2-2-2024-1/compspec
chmod +x compspec

# Download the spec for our compatibility artifact
wget --quiet https://gist.githubusercontent.com/vsoch/fcd0f7d633860674cb085a8540ce4bb2/raw/6afeaad2f941414fdb744788643c8b3566c6c531/lammps-experiment.yaml

# Generate!
./compspec create --in ./lammps-experiment.yaml -a custom.gpu.available=$hasGpu -o ${path}
cat ${path}
