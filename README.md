# Teapot [![Build Status](https://travis-ci.org/luan/teapot.svg?branch=master)](https://travis-ci.org/luan/teapot)

Teapot is the API that powers [Tiego](https://github.com/luan/tiego). It's a simple, RESTful JSON API to manage cloud provided workstations backed by [Diego](https://github.com/cloudfoundry-incubator/diego-release) and [Docker](http://docker.com) images.

You can find the live API documentation here: http://docs.teapot.apiary.io/

## Setup

Just go get it: `go get github.com/luan/teapot`

## Usage

To run the Teapot API with [Diego Edge](https://github.com/pivotal-cf-experimental/diego-edge):

```bash
teapot -address 0.0.0.0:8080 -receptorAddress http://receptor.192.168.11.11.xip.io/
```

To run the Teapot API **on** [Diego Edge](https://github.com/pivotal-cf-experimental/diego-edge):

```bash
cd $GOPATH/src/github.com/luan/teapot
export RECEPTOR=http://receptor.192.168.11.11.xip.io/
bin/dance # optionally provide a bucket name, default: tiego-artifacts
```

You can replace the receptor URL with your receptor for other types of deployments, receptors with Basic Auth enabled should work out of the box with something like `http://user:password@receptor.example.com`.

## Development flow

To deploy the Teapot to a Diego, we use a [minimal busybox image](https://github.com/jpetazzo/docker-busybox/blob/4f6cb64c3b3255c58021dc75100da0088796a108/Dockerfile) and download the compiled binary for Teapot and the [spy](https://github.com/cloudfoundry-incubator/docker-circus/tree/master/spy) from the [docker-circus](https://github.com/cloudfoundry-incubator/docker-circus).

The Teapot binary is stored on an S3 bucket, so before you deploy your changes to your Diego environment you need to get s3 working.

Also, if you're not developing from a linux workstation, you will need golang with [cross-compile support](#golang-cross-compile-on-osx).


### Install & Configure s3cmd
```bash
# for OSX
brew install s3cmd
# for debian based linux distros
sudo apt-get install -y s3cmd
s3cmd --configure # enter credentials Amazon S3 enabled account
```

### Build & Upload

```bash
bin/build
bin/upload # optionally provide a bucket name, default: tiego-artifacts
bin/deploy
```

Or you can use `bin/dance [BUCKET]` for convenicence

After this you can deploy Teapot using `bin/deploy`, as explained above.

### Golang cross-compile on OSX:

```bash
brew install go --cross-compile-all # or reinstall if you already had it installed
```
