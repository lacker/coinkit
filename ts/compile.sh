#!/bin/bash

tsc --build tsconfig.node.json || exit 1
echo node compile ok
tsc --build tsconfig.browser.json || exit 1
echo browser compile ok
