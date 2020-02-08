// Package parser parses the following grammar:
//
//
//     expr = app
//          | op
//          | unit
//          ;
//
//     op = expr id expr
//        ;
//
//     app = id expr ( expr ) *
//         ;
//
//     unit = group
//          | num
//          | arr
//          | id
//          ;
//
//     arr = ( num ) *
//         ;
//
//     group = "(" expr ")"
//           ;
//
//     id = ?? valid identifier characters ??
//
//     num = ?? valid number characters ??
//
//
// Both `app` and `op` are context-sensitive. With the use of a reference to
// the local environemnt, the parser can tell if it should continue parsing a
// function application with a specific number of arguments or an infix
// operator.
package parser
