// mmverify.py -- Proof verifier for the Metamath language
// Copyright (C) 2002 Raph Levien raph (at) acm (dot) org
// Copyright (C) David A. Wheeler and mmverify.py contributors
//
// This program is free software distributed under the MIT license;
// see the file LICENSE for full license information.
// SPDX-License-Identifier: MIT
//
// To run the program, type
//
//	$ python3 mmverify.py set.mm --logfile set.log
//
// and set.log will have the verification results.  One can also use bash
// redirections and type '$ python3 mmverify.py < set.mm 2> set.log' but this
// would fail in case 'set.mm' contains (directly or not) a recursive inclusion
// statement $[ set.mm $] .
//
// To get help on the program usage, type
//
//	$ python3 mmverify.py -h
//
// (nm 27-Jun-2005) mmverify.py requires that a $f hypothesis must not occur
// after a $e hypothesis in the same scope, even though this is allowed by
// the Metamath spec.  This is not a serious limitation since it can be
// met by rearranging the hypothesis order.
// (rl 2-Oct-2006) removed extraneous line found by Jason Orendorff
// (sf 27-Jan-2013) ported to Python 3, added support for compressed proofs
// and file inclusion
// (bj 3-Apr-2022) streamlined code; obtained significant speedup (4x on set.mm)
// by verifying compressed proofs without converting them to normal proof format;
// added type hints
// (am 29-May-2023) added typeguards
// (gn 23-Mar-2025) ported to Go.
package core

func main() {
	Vprint(1, "mmverify.go -- port of mmverifier.py")
	mm := NewMM(nil, nil)
	dbFile := os.Args[1]
	toks, err := NewToks(dbFile, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
	err = mm.Read(toks)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}
