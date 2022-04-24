#!/usr/bin/env bash

echo expect 401
curl -i -X POST http://localhost:8080/set/hi/item/x
echo

echo expect 401
curl -i -H "ttlset-auth: wrong" -X POST http://localhost:8080/set/hi/item/x
echo

echo expect 403
curl -i -H "ttlset-auth: key2" -X POST http://localhost:8080/set/hi/item/x
echo

echo expect 200
curl -i -H "ttlset-auth: key1" -X POST http://localhost:8080/set/hi/item/x
echo

echo expect 200 and count=1
curl -i -H "ttlset-auth: key2" http://localhost:8080/set/hi/count
echo
