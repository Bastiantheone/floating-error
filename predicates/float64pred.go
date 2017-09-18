package predicates

import (
	"math"
)

const (
	// macheps is the machine epsilon aka unit roundoff
	// The machine epsilon is an upper bound on the absolute relative true error in
	// representing a number.
	// If y is the machine representation of x then |(x-y)/x| <= macheps
	// https://en.wikipedia.org/wiki/Machine_epsilon
	// Go's float64 type has a 52-bit fractional mantissa,
	// therefore the value 2^-52
	macheps = 1.0 / (1 << 52)
)

// Float64Pred dynamically updates the potential error.
//
// If y is the machine representation of x then |(x-y)/x| <= macheps and |x-y| = e.
// Since we want the max possible error we assume |(x-y)/x| = macheps
// macheps*|x| = |x - y|
// if (x-y)>=0 then macheps*|x| = x - y  ->  y = x- macheps*|x|
// else macheps*|x| = -(x - y)  ->  macheps*|x|+x = y
// Each one of these has two cases again. x can be positive or negative, resulting in the same two possible equations:
// y/(1-macheps)=x or y/(1+macheps)=x.
// Since x is unknown, we will use the larger number as factor, to avoid having an error greater than
// the maxError value we have.
// |y-x| = macheps*|x| -> e = macheps*|x| -> e = macheps* |y|/(1-macheps)
// A Special case is when y = 0. Then we use the smallest nonzero float, because that is the max
// possible error in this case.
type Float64Pred struct {
	// n is the number
	n float64
	// e is the max rounding error possible
	e float64
}

// NewFloat64Pred returns a new float64Pred e set to 0.
func NewFloat64Pred(n float64) Float64Pred {
	return Float64Pred{
		n: n,
		e: 0,
	}
}

// GetValues returns the number and the potential error stored in p.
func (p Float64Pred) GetValues() (number, error float64) {
	return p.n, p.e
}

// AddFloat64 adds f to p and updates the potential error
func (p Float64Pred) AddFloat64(f float64) Float64Pred {
	p.n += f
	if p.n == 0 {
		p.e += math.SmallestNonzeroFloat64
	} else {
		p.e += macheps * math.Abs(p.n) / (1 - macheps)
	}
	return p
}

// AddFloat64Pred adds b to a and updates the potential error
func (a Float64Pred) AddFloat64Pred(b Float64Pred) Float64Pred {
	a.n += b.n
	if a.n == 0 {
		a.e += math.SmallestNonzeroFloat64 + b.e
	} else {
		a.e += macheps*math.Abs(a.n)/(1-macheps) + b.e
	}
	return a
}

// SubFloat64 subtracts f from p and updates the potential error
func (p Float64Pred) SubFloat64(f float64) Float64Pred {
	p.n -= f
	if p.n == 0 {
		p.e += math.SmallestNonzeroFloat64
	} else {
		p.e += macheps * math.Abs(p.n) / (1 - macheps)
	}
	return p
}

// SubFloat64Pred subtracts a from b and updates the potential error
func (a Float64Pred) SubFloat64Pred(b Float64Pred) Float64Pred {
	a.n -= b.n
	if a.n == 0 {
		a.e += math.SmallestNonzeroFloat64 + b.e
	} else {
		a.e += macheps*math.Abs(a.n)/(1-macheps) + b.e
	}
	return a
}

// MulFloat64 multiplies p with f and updates the potential error.
//
// mul(mul(a,b),c) = mul(a*b+error,c) = a*b*c + error*c + newError
// sum(mul(a,b),c) = sum(a*b+error,c) = a*b+c + error + newError
// Conclusively, when multiplications are chained, the error also depends on the value
// of the number, but this does not apply to sums or subtractions.
//
//If this is not a chained multiplication p.e will be 0, making that part irrelevant.
func (p Float64Pred) MulFloat64Pred(f float64) Float64Pred {
	p.n *= f
	if p.n == 0 {
		p.e += math.SmallestNonzeroFloat64 + p.e*math.Abs(f)
	} else {
		p.e += macheps*math.Abs(p.n)/(1-macheps) + p.e*math.Abs(f)
	}
	return p
}

// MulFloat64Pred multiplies a with b and updates the potential error.
//
// mul(mul(a,b),c) = mul(a*b+error,c) = a*b*c + error*c + newError
// sum(mul(a,b),c) = sum(a*b+error,c) = a*b+c + error + newError
// Conclusively, when multiplications are chained, the error also depends on the value
// of the number, but this does not apply to sums or subtractions.
//
// mul(sum(a,b),sum(c,d)) = mul(a+b+e1,c+d+e2) = (a+b)*(c+d) + e1*(c+d+e2) + e2*(a+b+e1) + newError
//
//If this is not a chained multiplication a.e/b.e will be 0, making that part irrelevant.
func (a Float64Pred) MulFloat64(b Float64Pred) Float64Pred {
	a.n *= b.n
	if a.n == 0 {
		a.e += math.SmallestNonzeroFloat64 + a.e*math.Abs(b.n) + b.e*math.Abs(a.n)
	} else {
		a.e += macheps*math.Abs(a.n)/(1-macheps) + a.e*math.Abs(b.n) + b.e*math.Abs(a.n)
	}
	return a
}
