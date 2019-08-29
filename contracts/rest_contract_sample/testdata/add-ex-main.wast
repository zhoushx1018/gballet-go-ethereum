;; a module that imports two other modules, and defines 3 functions
;; using exported functions from these two imported modules.
;;
;; see https://github.com/WebAssembly/spec/tree/master/interpreter/#s-expression-syntax
;; for a reference about the syntax.
(module
  (memory (export "memory") 1)
  (func (export "main")
  )
)