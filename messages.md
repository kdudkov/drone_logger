# Messages from drone

## 0x8a - ?
`58 8a 0c 584c30313253465878800c08 10`

## 0x8b - coords
`58 8b 0e 3900 f300 7e3d fb00290000000007 d9`

* 1 `58` header
* 2 `8b` message type
* 3 `0e` payload len
* 4-5 `3900` roll*100 (-180 to +180)
* 6-7 `f300` pitch*100 (-180 to +180)
* 8-9 `7e3d` yaw*100 (-180 to +180)
* 10-17 `fb00290000000007` - ?
* 18 `d9` checksum

## 0x8c - coords
`58 8c 0d 00000000 00000000 e803000000 6a`
* 1 `58` header
* 2 `8c` message type
* 3 `0d` payload len
* 4-7 `00000000` longitude*10000000 32bit signed int (LE)
* 8-11 `00000000` latitude*10000000 32bit signed int (LE)
* 12-16 `e803000000` - ??
* 17 `6a` checksum

## 0x8f
`58 8f 08 c0 f00c0000010000 ba`
* 1 `58` header
* 2 `8f` message type
* 3 `08` payload len
* 4 `c0` electromagnetic field
* 5-11 `f00c0000010000` - ?
* 12 `ba` checksum

## 0xfe - some text
`58 fe 19 48462d584c2d584c303132533031312e3032312e3132323205 8c`
* 1 `58` header
* 2 `fe` message type
* 3 `19` payload len
* 4-(n-1) text (HF-XL-XL012S011.021.1222)
* n `8c` checksum


# commands to drone
(only code and payload)

* `01` `80808080200800000100000000`
* `7e` `48462d58582d5858585858583030302e3030302e30303030000000000000`
* `0b` `0000000000000000`
* `1b` `0000000000000000`
* `0a` `000000000000000000000000000000`