#! /bin/bash
set -e #u

if [ "$#" -ne 1 ]; then echo "Usage: ./upload_all_images VERSION_TAG"; exit; fi

VERSION=$1
ACCOUNT=001138754299 #$(aws sts get-caller-identity --query Account --output text)
REGION=us-west-2

aws ecr get-login-password --region us-west-2 --profile Constellation-Admin/PowerUser | docker login --username AWS --password-stdin $ACCOUNT.dkr.ecr.us-west-2.amazonaws.com

build_tag_push () {
  echo $1
  docker build -t $1 -f Dockerfile $2 &&
  docker tag $1 $ACCOUNT.dkr.ecr.us-west-2.amazonaws.com/$1:$VERSION &&
  docker image push $ACCOUNT.dkr.ecr.us-west-2.amazonaws.com/$1:$VERSION
}
build_tag_push op-plasma-celestia .
