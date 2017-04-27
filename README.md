# docksible

Build [docker](http://www.docker.com) images with [ansible](http://www.ansible.com).

:cactus: First release available for downloading in [releases](https://github.com/localghost/docksible/releases/tag/v0.1.0).

## What is docksible?

Docksible is a utility tool that helps you build docker images provisioned with ansible's playbooks.

## Why docksible?

There are other ways to provision docker containers with ansible, like simply using `ansible-playbook` with `ansible_connection` set to `docker` or using [`ansible-container`](https://github.com/ansible/ansible-container). However, at least for me, neither of the available solutions was convenient enough to use. They either required some extra glue code or flooded my machine with multiple dependencies. While what I wanted was to have a single binary through which I could provision my containers using exactly the same playbooks I use for VMs or bare metal hosts. And so `docksible` was born.

Ahh, and I also wanted to learn [Go](https://golang.org/).

## Requirements

The only requirements for docksible are: docksible binary, docker engine and a playbook to run.

## How to use it

Assuming you have following structure of ansible's scripts:
```
ansible/
  - playbook.yml
  - roles/*
```

you enter `ansible` directory and you run:
```
docksible --result-image my_image centos:7.3.1611 playbook.yml
```

And soon you have image `my_image` based on CentOS 7.3 ready waiting for you in your docker engine.

[![Build Status](https://travis-ci.org/localghost/docksible.svg?branch=master)](https://travis-ci.org/localghost/docksible)
