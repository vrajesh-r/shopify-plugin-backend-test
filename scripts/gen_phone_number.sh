#!/bin/sh
echo generating a fake phone number for checkout!
phone='000'
for i in {1..7}
do
    phone="${phone}"$(($RANDOM%10))
done
echo $phone | pbcopy
echo $phone is on your clipboard now!