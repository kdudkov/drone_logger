# Messages from drone

* 1 `58` header
* 2 message type
* 3 payload len (message len - 4)
* 4-(n-1) payload
* n - checksum (xor of bytes 2-(n-1))

## 0x8a - ?

`584c30313253465878 800c08`

* 1-9 text XL012SFXx
* 10-12 `800c08`

## 0x8b - gyro

`3900 f300 7e3d fb00290000000007`

* 1-2 `3900` roll*100 (-180 to +180)
* 3-4 `f300` pitch*100 (-180 to +180)
* 5-6 `7e3d` yaw*100 (-180 to +180)
* 7-14 `fb00290000000007` - ?

## 0x8c - coords

`00000000 00000000 e803 0000`

* 1-4 `00000000` longitude*10000000 32bit signed int (LE)
* 5-8 `00000000` latitude*10000000 32bit signed int (LE)
* 9-13 `e803000000` - ??

## 0x8f

`c0 f00c 0000010000`

* 1 `c0` electromagnetic field
* 2-3 `f00c` battery voltage * 400
* 4-8 `0000010000` - ?

## 0xfe - some text

`48462d584c2d584c303132533031312e3032312e31323232`

* text (HF-XL-XL012S011.021.1222)

# commands to drone

* 1 `68` header
* 2 message type
* 3 payload len (message len - 4)
* 4-(n-1) payload
* n - checksum (xor of bytes 2-(n-1))

## 0x01 - control

* `01` `80808080200800000100000000`

## other

* `7e` `48462d58582d5858585858583030302e3030302e30303030000000000000`
  -> answer `0xfe`
* `0b` `0000000000000000`
* `1b` `0000000000000000`
* `0a` `000000000000000000000000000000` -> answer `0x8a`