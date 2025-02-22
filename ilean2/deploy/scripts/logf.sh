#!/usr/bin/env bash

services="$services -u lean"

journalctl $lean -f
