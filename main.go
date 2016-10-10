package main

import _ "net/http/pprof"

import (
	"bytes"
	"fmt"
	"github.com/pkg/profile"
	// "log"
	"math/big"
	// "net/http"
	"time"
)

type exp_fn func(*big.Int, *big.Int, *big.Int) *big.Int

const (
	HexBase = 16
	K       = 2048
)

var zero = big.NewInt(0)
var one = big.NewInt(1)

var e = new(big.Int)
var d = new(big.Int)
var n = new(big.Int)
var n_prime = new(big.Int)
var r = new(big.Int).Lsh(big.NewInt(1), K)
var r_ = new(big.Int).Sub(r, one)
var r_inv = new(big.Int)

func copy_big(a *big.Int) *big.Int {
	return new(big.Int).Add(a, zero)
}

func pow_mod(a, b, c *big.Int) *big.Int {
	ret := big.NewInt(1)
	tmp := new(big.Int)
	_a := copy_big(a)
	_b := copy_big(b)
	_c := copy_big(c)

	for _b.Cmp(zero) != 0 {
		if tmp.And(_b, one).Cmp(zero) != 0 {
			ret.Mul(ret, _a).Mod(ret, _c)
		}
		_b.Rsh(_b, 1)
		_a.Mul(_a, _a).Mod(_a, _c)
	}

	return ret
}

func extended_gcd(a, b *big.Int) (*big.Int, *big.Int) {
	s, old_s := big.NewInt(0), big.NewInt(1)
	t, old_t := big.NewInt(1), big.NewInt(0)
	r, old_r := copy_big(b), copy_big(a)
	q := new(big.Int)
	tmp := new(big.Int)

	for r.Cmp(zero) != 0 {
		q.Div(old_r, r)

		tmp.Mul(q, r)
		old_r, r = r, old_r
		r.Sub(r, tmp)

		tmp.Mul(q, s)
		old_s, s = s, old_s
		s.Sub(s, tmp)

		tmp.Mul(q, t)
		old_t, t = t, old_t
		t.Sub(t, tmp)
	}

	if old_s.Cmp(zero) < 0 && old_t.Cmp(zero) > 0 {
		old_s.Add(old_s, b)
		old_t.Sub(old_t, a)
	}

	return old_s, old_t
}

func mon_pro(a, b *big.Int) *big.Int {
	t := new(big.Int).Mul(a, b)
	m := new(big.Int).And(new(big.Int).Mul(new(big.Int).And(t, r_), n_prime), r_)
	u := new(big.Int).Add(t, new(big.Int).Mul(m, n))
	u.Rsh(u, K)

	if u.Cmp(n) >= 0 {
		u.Sub(u, n)
	}

	return u
}

func mod_exp(a, e, n *big.Int) *big.Int {
	r_inv, n_prime = extended_gcd(r, n)
	n_prime.Neg(n_prime)
	n_prime.And(n_prime, r_)

	a_bar := new(big.Int).Mul(a, r)
	a_bar.Mod(a_bar, n)

	x_bar := new(big.Int).Mod(r, n)

	for i := e.BitLen(); i >= 0; i-- {
		x_bar = mon_pro(x_bar, x_bar)
		if e.Bit(i) == 1 {
			x_bar = mon_pro(a_bar, x_bar)
		}
	}

	return mon_pro(x_bar, one)
}

func sign(m string, pow_fn exp_fn) string {
	z := new(big.Int).SetBytes([]byte(m))
	// z.Exp(z, d, n)
	z = pow_fn(z, d, n)

	return z.Text(HexBase)
}

func verify(m, s string, pow_fn exp_fn) bool {
	z, success := new(big.Int).SetString(s, HexBase)
	if !success {
		fmt.Println(success)
	}
	// z.Exp(z, e, n)
	z = pow_fn(z, e, n)

	// fmt.Printf("%q\n", m1.Bytes())
	return bytes.Equal([]byte(m), z.Bytes())
}

func main() {
	defer profile.Start().Stop()

	// Randomly generated using: http://www.mobilefish.com/services/rsa_key_generation/rsa_key_generation.php
	e.SetString("10001", HexBase)
	d.SetString("eb865f1cbcbcfdec6b693be044e8338e35349352d3599bddf4698572d2618fb9e8d2b15be2807b603d030a53ee454535b020276bc1632c791ef52dc44e1418ac6c2e668ef102dc6f33063795b1b291cd5ecaac80d092bbab6ef6d6faac34e621ee223bc8a3b0c50f7b0025bfd60eaf763831edc22eb9230617c5e64f370384c59791a11b4fad2ebb441ecfdbba67e42b35100fc100fdc97434944dad923465a1c238488735178eb474b04850652d6703103e27a9816350f313251ac847cba2ac26b5104d988e7f0ab10deebffe5d69c9bbaffde39fbdbe201372c15d0631ba9e9c84e0f1a616180b4ddc07efd145e9cbfbc36910f4dc04463f6f7edf732af031", HexBase)
	n.SetString("f18ff84197460d9e7fdd494f48c4cadb96eeee61bdae5b245fe090da9dc74d8e682cc0588a06fc8dae5f4ec9c8eafa0be35aaa4ef3ab12cb7a9528859a2b3d3f29c0d0b3e5ef1f86a7829081f8618b3f5cc43e2d13500b15081f3582afde29f93afa4c75ccbfae76de2a450b7e4d28eb9204df1ac299b2921b131f5ca8d65e95d57101d1f250070c9f10d84330e3f7775d51a9e65106845251c59577415168433ceccbcc8cbabf9d51a8bbff0901fd26261bf5eba8b8ead797266d8ce7d7097adb9d5296482eced88bfc70ae0a62bf4eb35d861297ed46926fd971d9c9f9d9e655ad16b58270238eee17afd78c3765aa0a67dae01afc782b31dce1c31fef42fb", HexBase)

	m := "hello world"
	// m1 := "dlrow olleh"

	N := 1000
	var t time.Time

	t = time.Now()
	for i := 0; i < N; i++ {
		s := sign(m, mod_exp)
		verify(m, s, mod_exp)
	}
	fmt.Println(time.Since(t))

	t = time.Now()
	for i := 0; i < N; i++ {
		s := sign(m, pow_mod)
		verify(m, s, pow_mod)
	}
	fmt.Println(time.Since(t))

	// s := sign(m, mod_exp)
	// fmt.Println(verify(m, s, mod_exp))
	// fmt.Println(verify(m1, s, mod_exp))
}
