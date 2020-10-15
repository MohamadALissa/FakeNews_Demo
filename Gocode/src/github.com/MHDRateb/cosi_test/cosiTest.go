package main
// #include <stdio.h>
// #include <stdlib.h>
//
import "C"

import (
    "fmt"
    "crypto/cipher"
    "crypto/sha512"
    "hash"

    // "os"
    // "strconv"
    "go.dedis.ch/kyber"

    "go.dedis.ch/kyber/sign/cosi"
    "go.dedis.ch/kyber/suites"
    "go.dedis.ch/kyber/group/edwards25519"

    "go.dedis.ch/kyber/sign/eddsa"
    "go.dedis.ch/kyber/util/key"
    "go.dedis.ch/kyber/xof/blake2xb"
    "go.dedis.ch/kyber/util/random"

    // "unsafe"
)



type cosiSuite struct {
    suites.Suite
    r kyber.XOF
}

func (m *cosiSuite) Hash() hash.Hash {
    return sha512.New()
}
func (m *cosiSuite) RandomStream() cipher.Stream { return m.r }


func main() {
    

}

//export startSign
func startSign(orginalMessage string) *C.char {

    m:=orginalMessage

    argCount := len(orginalMessage)
    
        if (argCount>0) {m= string(orginalMessage)}
    
    
    fmt.Printf("Message to sign: %s\n\n",m)
    message := []byte(m)
    sigaggr :=testCoSi(message,3, 0)

    finalstring := C.CString(sigaggr)

    return finalstring   

}

type scalar struct {
	v [32]byte
}

func testCoSi(message []byte, n, f int) string {
    
    testSuite := &cosiSuite{edwards25519.NewBlakeSHA256Ed25519(), blake2xb.New(nil)}

    // Generate key pairs
    var kps []*key.Pair
    var privates []kyber.Scalar
    var publics []kyber.Point

    for i := 0; i < n; i++ {
            kp := key.NewKeyPair(testSuite)
    kp.Private = testSuite.Scalar().Pick(random.New())

    kp.Public = testSuite.Point().Mul(kp.Private, nil)

            kps = append(kps, kp)

            privates = append(privates, kp.Private)
            publics = append(publics, kp.Public)
        }

    fmt.Printf("Private keys: %s\n\n",privates)
    fmt.Printf("Public keys: %s\n\n",publics)


    // Init masks
    var masks []*cosi.Mask
    var byteMasks [][]byte
    for i := 0; i < n-f; i++ {
        m, _ := cosi.NewMask(testSuite, publics, publics[i])

        masks = append(masks, m)
        byteMasks = append(byteMasks, masks[i].Mask())
    }
    fmt.Printf("Masks: %x\n\n",masks)
    fmt.Printf("Byte masks: %x\n\n",byteMasks)


    // Compute commitments
    var v []kyber.Scalar // random
    var V []kyber.Point  // commitment
    for i := 0; i < n-f; i++ {
        x, X := cosi.Commit(testSuite)
        v = append(v, x)
        V = append(V, X)
    }

    fmt.Printf("Commitments: %s\n\n",V)

    // Aggregate commitments
    aggV, aggMask, _ := cosi.AggregateCommitments(testSuite, V, byteMasks)
    
    fmt.Printf("Aggregated commitment (V): %s\n\n",aggV)
    fmt.Printf("Aggregated mask (Z): %x\n\n",aggMask)

    // Set aggregate mask in nodes
    for i := 0; i < n-f; i++ {
        masks[i].SetMask(aggMask)
    }

    // Compute challenge
    var c []kyber.Scalar
    for i := 0; i < n-f; i++ {
        ci, _ := cosi.Challenge(testSuite, aggV, masks[i].AggregatePublic, message)
        
        c = append(c, ci)
    }

    fmt.Printf("Challenge: %s\n\n",c)

    // Compute responses
    var r []kyber.Scalar
    for i := 0; i < n-f; i++ {
        ri, _ := cosi.Response(testSuite, privates[i], v[i], c[i])
        r = append(r, ri)
    }
    fmt.Printf("Responses: %s\n\n",r)

    // Aggregate responses
    aggr, _ := cosi.AggregateResponses(testSuite, r)

    fmt.Printf("Aggregated responses (r): %s\n\n",aggr)

    for i := 0; i < n-f; i++ {
        // Sign
        sig, _ := cosi.Sign(testSuite, aggV, aggr, masks[i])
        fmt.Printf("Signature (%d): %x",i,sig)


        // Set policy depending on threshold f and then Verify
        var p cosi.Policy
        if f == 0 {
            p = nil
        } else {
            p = cosi.NewThresholdPolicy(n - f)
        }
        // send a short sig in, expect an error
        if err := cosi.Verify(testSuite, publics, message, sig[0:10], p); err == nil {
            
        }
        if err := cosi.Verify(testSuite, publics, message, sig, p); err != nil {
            
        }

        // cosi signature should follow the same format as EdDSA except it has no mask
        maskLen := len(masks[i].Mask())

        if err := eddsa.Verify(masks[i].AggregatePublic, message, sig[0:len(sig)-maskLen]); err != nil {
            
        } else {
            fmt.Printf("..Verified.\n")
        }   

    }

    sigaggr :=kyber.Scalar.String(aggr)
    return sigaggr

}
