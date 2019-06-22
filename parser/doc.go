// Package parser parses the following grammar:
//
// expr = app
//      | op
//      | unit
//      ;
//
// op = expr id expr
//    ;
//
// app = id expr ( expr ) *
//     ;
//
// unit = group
//      | num
//      | id
//      ;
//
// group = "(" expr ")"
//       ;
//
// id = ?? valid identifier characters ??
//
// num = ?? valid number characters ??
package parser
