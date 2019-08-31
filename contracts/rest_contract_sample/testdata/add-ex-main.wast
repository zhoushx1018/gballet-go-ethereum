
(module
    (import "debug" "printMemHex" (func $printMemHex (param i32 i32)))
    (import "debug" "printStorageHex" (func $printStorageHex (param i32 )))
    (import "ethereum" "getCallDataSize" (func  $getCallDataSize (result i32)))
    (import "ethereum" "storageStore" (func $storageStore (param i32 i32)))
    (import "ethereum" "getBlockDifficulty" (func $getBlockDifficulty(param i32)))
    (memory 1)
    (export "memory" (memory 0))
    (export "main" (func $main))
    (func $main

      (call $printMemHex (i32.const 0) (i32.const 32))
      (i32.store (i32.const 0) (i32.const 100))
      (call $printMemHex (i32.const 0) (i32.const 32))

;;      (call $getBlockDifficulty(i32.const 11))
;;      (call $printStorageHex (i32.const 11))
;;      (call $storageStore (i32.const 0) (i32.const 0))
;;      (call $printStorageHex (i32.const 100))
;;      (call $storageStore (i32.const 100) (i32.const 0))
;;      (call $printStorageHex (i32.const 100))

    )
)
