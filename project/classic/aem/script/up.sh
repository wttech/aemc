#!/usr/bin/env sh

step "creating instances"
aem instance create
clc

step "starting instances (it may take a few minutes)"
aem instance start
clc
