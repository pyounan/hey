# POS-Proxy




## FDM Constraints

### LineItem

Each item entry length should be exactly 33 letters,

item = quantity + description + price + VAT code


#### Item Description

Description length should be exactly 20, if length is smaller than 20, append spaces to the right.


#### Item Price

 1. Price length should be exactly 8, if length is smaller than 8, prepend zeros.
 2. The last two letters are the letters after the decimal point.


#### Item Quantity

 1. Quantity length should be 4 letters, if len is smaller than 4, prepend zeros.
 2. Quantity should be in grams (not kg), and milliliters (not liters)... etc

### Hash And Sign

Format: identifier + sequence + retry + ticket_date + ticket_time_period + user_id + RCRS + ticket_number + event_label + total_amount + 4 vats + plu

#### Ticket Number

length: 6 

if length is less than 6, fill with spaces from the left

#### Total Amount

length: 11

if length is less than 11, fill with spaces from the left,
that last two digits are always the number after the decimal point.

#### VAT Codes

A
B
C
D

#### UserID

Length: 14
Social Security Number


#### Encryption & Decryption of auth key
https://astaxie.gitbooks.io/build-web-application-with-golang/en/09.6.html
