package peer

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	cr "modubft/crypto"
	pb "modubft/proto"
	"os"

	"google.golang.org/protobuf/proto"
)

func generateMACKeys(peerID int, ids []int) (MACKeys [][]byte, MACCipher []cipher.AEAD, MACNonce [][]byte) {

	idListLength := len(ids)
	MACKeys = make([][]byte, idListLength)
	MACCipher = make([]cipher.AEAD, idListLength)
	MACNonce = make([][]byte, idListLength)

	for _, id := range ids {
		basePhrase := []byte("passphrasewhichneedstobe32bytes!")
		basePhrase[0] = byte(peerID + int(id))
		MACKeys[id] = basePhrase
	}

	for _, id := range ids {
		cypher, _ := aes.NewCipher(MACKeys[id])
		gcm, _ := cipher.NewGCM(cypher)
		MACCipher[id] = gcm

		nonce := make([]byte, gcm.NonceSize())
		for x := 0; x < gcm.NonceSize(); x++ {
			nonce[x] = byte(peerID + int(id))
		}
		MACNonce[id] = nonce
	}

	return
}

func generatePublicKeys(ids []int) (publicKeys []string, pkVerifier []cr.Unsigner) {
	publicKeys = make([]string, len(ids))
	pkVerifier = make([]cr.Unsigner, len(ids))
	for _, id := range ids {
		publicKeys[id] = fmt.Sprintf("ID%d", id)
	}
	return
}

// VerifySig Verify the validity of a received message
func (p *Node) VerifySig(pm *pb.PeerMessage) bool {

	if pm.Auth == nil {
		fmt.Println("Message without Auth from ", pm.FromNodeId)
		os.Exit(0)
		return true
	}

	success := false

	if pm.FromNodeId == int32(p.myID) {
		p.uncompressAndValidateCertificates(pm, true)
		return true
	}

	msgAuth := pm.Auth
	pm.Auth = nil

	msgHash := p.computeMsgHash(pm)

	pm.Auth = msgAuth

	for _, elem := range msgAuth {
		if elem.ToNodeId == int32(p.myID) || elem.ToNodeId == -1 {
			// found a signature for me

			if elem.SigType == pb.SigType_MAC {
				// start := time.Now()
				success = p.verifyMACSig(elem.FromNodeId, msgHash, elem.Sig)
				// p.incrementTimeSpent(fmt.Sprintf("verifyMACSig:%s", pm.Type.String()), float64(time.Since(start).Nanoseconds()))
			} else if elem.SigType == pb.SigType_PKSig {
				// start := time.Now()
				success = p.verifyPKSig(elem.FromNodeId, msgHash, elem.Sig)
				// p.incrementTimeSpent(fmt.Sprintf("verifyPKSig:%s", pm.Type.String()), float64(time.Since(start).Nanoseconds()))
			}

			break
		}
	}

	if success {

		success = p.uncompressAndValidateCertificates(pm, success)

	}

	return success
}

func (p *Node) uncompressAndValidateCertificates(pm *pb.PeerMessage, success bool) bool {
	if p.useCryptoForCertificates && ((!p.ClusterPrepareToAllAndEnableHashForClusterCommit && pm.Type == pb.MsgType_ClusterCommit) || (p.ClusterPrepareToAllAndEnableHashForClusterCommit && pm.Type == pb.MsgType_ClusterPrepare)) {
		if p.useCertificateCompression && pm.Certificate == nil && pm.CompressedCertificates != nil && len(pm.CompressedCertificates) > 0 {
			// fmt.Printf("Node %d: Certificate %v and compressed certificates %v, MsgId=%d \n", p.myID, pm.Certificate, pm.CompressedCertificates, pm.MsgId)

			// fmt.Printf("Node %d: Received ClusterCommit without certificates, MsgId=%d msg:%+v \n", p.myID, pm.MsgId, pm)

			r, _ := zlib.NewReader(bytes.NewReader(pm.CompressedCertificates))
			reader := bufio.NewReader(r)
			pm.Certificate = make([]*pb.ClusterCertificate, 0)
			uncrompressed := make([]byte, 0)
			for {
				b, err := reader.ReadByte()
				if err != nil {
					if err == io.EOF {
						break
					} else {
						panic(err)
					}
				}

				uncrompressed = append(uncrompressed, b)

			}
			r.Close()

			// fmt.Printf("Node %d: Uncompressed ClusterCommit certificate size bytes: %d\n", p.myID, len(uncrompressed))
			// fmt.Printf("Node %d: Uncompressed ClusterCommit certificate bytes: %x\n", p.myID, uncrompressed)
			i := 0
			for i < len(uncrompressed) {
				sizeBytes := uncrompressed[i : i+4]
				// fmt.Printf("Node %d: Uncompressed ClusterCommit certificate size bytes: %x\n", p.myID, sizeBytes)
				i += 4

				pbsize := int(binary.LittleEndian.Uint32(sizeBytes))
				// fmt.Printf("Node %d: Uncompressed ClusterCommit certificate size: %d\n", p.myID, pbsize)
				pbbytes := make([]byte, pbsize)
				copy(pbbytes, uncrompressed[i:i+pbsize])
				// fmt.Printf("Node %d: Uncompressed ClusterCommit certificate bytes: %x\n", p.myID, pbbytes)
				i = i + pbsize
				cert := pb.ClusterCertificate{}
				err := proto.Unmarshal(pbbytes, &cert)
				if err != nil {
					panic(err)
				}
				pm.Certificate = append(pm.Certificate, proto.Clone(&cert).(*pb.ClusterCertificate))
				// fmt.Printf("Node %d: Uncompressed ClusterCommit certificate %d: %+v\n", p.myID, len(pm.Certificate), pm.Certificate)
			}

			// pm.Certificate = result

			// fmt.Printf("Node %d: Certificate %v and compressed certificates %v, MsgId=%d \n", p.myID, pm.Certificate, pm.CompressedCertificates, pm.MsgId)

			// fmt.Printf("Node %d: Uncompressed ClusterCommit certificates, MsgId=%d, Certificates=%+v msg:%+v \n", p.myID, pm.MsgId, pm.Certificate, pm)
		}

		success = pm.FromNodeId == int32(p.myID) || p.validateCertificates(pm)
	}
	return success
}

type ToHash struct {
	i      int
	toHash []byte
}
type HashOut struct {
	i    int
	hash [32]byte
}

func hash(toHash chan ToHash, hashOut chan HashOut) {
	for to := range toHash {
		result := HashOut{
			i:    to.i,
			hash: sha256.Sum256(to.toHash),
		}
		hashOut <- result
	}
}

func (p *Node) validateCertificates(pm *pb.PeerMessage) bool {
	if pm.MsgId == 0 {
		// fmt.Printf("Node %d: Validating certificate: MsgId=%d, FromNodeId=%d, Certificate=%x and SIG %x\n", p.myID, pm.MsgId, pm.FromNodeId, pm.Certificate, pm.Auth)
	}
	if len(pm.Certificate) < p.f {
		panic(fmt.Sprintf("Not enough Certificates: got %d, expected at least %d (f=%d), MsgId=%d", len(pm.Certificate), p.f, p.f, pm.MsgId))
		return false
	}

	// fmt.Printf("Node %d: Validating %d certificates for MsgId=%d\n", p.myID, len(pm.Certificate), pm.MsgId)

	// start := time.Now()
	state := p.generateStateFromCert(pm)
	firstCertHash := sha256.Sum256(state)
	// p.incrementTimeSpent(fmt.Sprintf("generateStateFromCert:%s", pm.Type.String()), float64(time.Since(start).Nanoseconds()))

	// firstCertHash := sha256.Sum256(pm.Certificate[0].Certificate)
	certsWithHashes := make([]*pb.ClusterCertificate, len(pm.Certificate))
	for i, cert := range pm.Certificate {
		// certHash := sha256.Sum256(cert.Certificate)
		certsWithHashes[i] = &pb.ClusterCertificate{
			FromNodeId: pm.FromNodeId,
		}
		// copy(certsWithHashes[i].Certificate, state[:])

		if pm.MsgId == 0 {
			// fmt.Printf("Verifying certificate: MsgId=%d, FromNodeId=%d, Certificate=%x and SIG %x\n", pm.MsgId, cert.FromNodeId, cert.Certificate, cert.Sig)
		}

		if cert.FromNodeId == int32(p.myID) {
			continue
		}

		// if !bytes.Equal(certHash[:], firstCertHash[:]) {
		// 	s := fmt.Sprintf("Node %d: Certificates are not equal! FromNodeId=%d, Certificate=%x, FirstCertificate=%x, MsgId=%d", p.myID, cert.FromNodeId, cert.Certificate, firstCert.Certificate, pm.MsgId)
		// 	for _, cert := range pm.Certificate {
		// 		s = s + fmt.Sprintf(" rest:  FromNodeId=%d, Certificate=%x", cert.FromNodeId, cert.Certificate)
		// 	}

		// 	panic(s + "\n")
		// 	return false
		// }
		// start = time.Now()
		if !p.verifyPKSig(cert.FromNodeId, firstCertHash, cert.Sig) {
			panic(fmt.Sprintf("Node %d: Error verifying PK signature from %d, MsgId=%d", p.myID, cert.FromNodeId, pm.MsgId))
			return false
		}
		// p.incrementTimeSpent(fmt.Sprintf("verifyCert:%s", pm.Type.String()), float64(time.Since(start).Nanoseconds()))

	}
	pm.AttachedData = state
	return true
}

func (p *Node) generateStateFromCert(pm *pb.PeerMessage) []byte {
	// return  pm.AttachedData
	channelsIn := make(chan ToHash, 100)
	channelsOut := make(chan HashOut, 100)

	// firstCert := pm.Certificate[0]

	// reader, _ := gzip.NewReader(bytes.NewReader(firstCert.Certificate))
	// data, _ := io.ReadAll(reader)
	data := pm.AttachedData

	state := make([]byte, p.ClusterBatchSize*32)
	returned := 0

	for range 8 {
		go hash(channelsIn, channelsOut)
	}
	go func() {
		start := 0
		for i := 0; i < p.ClusterBatchSize; i++ {
			currentOffset := start + int(pm.MessageSizes[i])
			newVar := data[start:currentOffset]
			channelsIn <- ToHash{i: i, toHash: newVar}
			start = currentOffset
		}
	}()

	for returned < p.ClusterBatchSize {
		hashed := <-channelsOut
		copy(state[hashed.i*32:], hashed.hash[:])
		returned++
	}

	close(channelsIn)
	close(channelsOut)
	// fmt.Printf("Returned %d hashes, cluster batch size %d\n", returned, p.ClusterBatchSize)

	return state
}

func (p *Node) verifyMACSig(fromId int32, hash [32]byte, sig []byte) bool {

	/*if p.MACCipher[fromId] == nil {
		cypher, _ := aes.NewCipher(p.MACKeys[fromId])
		gcm, _ := cipher.NewGCM(cypher)
		p.MACCipher[fromId] = gcm

		nonce := make([]byte, gcm.NonceSize())
		for i := 0; i < gcm.NonceSize(); i++ {
			nonce[i] = byte(p.myID + int(fromId))
		}
		p.MACNonce[fromId] = nonce
	}*/

	nonceSize := p.MACCipher[fromId].NonceSize()

	nonce, ciphertext := sig[:nonceSize], sig[nonceSize:]
	plaintext, err := p.MACCipher[fromId].Open(nil, nonce, ciphertext, nil)

	if err != nil {
		fmt.Printf("Failed to verifyMACSig(fromId %d, hash %x , sig %x): p.MACCipher %v \n ", fromId, hash, sig, p.MACCipher)
		panic(err)
		return false
	}

	return string(plaintext) == string(hash[:])

}

func (p *Node) createMACSig(toId int32, hash [32]byte) []byte {

	nonceSize := p.MACCipher[toId].NonceSize()

	mac := make([]byte, nonceSize)
	copy(mac, p.MACNonce[toId])
	mac = p.MACCipher[toId].Seal(mac, p.MACNonce[toId], hash[:], nil)

	return mac

}

func (p *Node) verifyPKSig(fromId int32, hash [32]byte, sig []byte) bool {
	// fmt.Printf("USING PK??\n")
	if p.PKVerifier[fromId] == nil {
		p.PKVerifier[fromId], _ = cr.LoadPublicKey(p.PubKey[fromId])
	}

	verifier := p.PKVerifier[fromId]

	err := verifier.Unsign(hash[:], sig)

	if err == nil {
		return true
	} else {

		panic(fmt.Sprintf("Node %d: Error verifying PK signature from %d: %s\n", p.myID, fromId, err))
		return false
	}
}

func (p *Node) createPKSig(hash [32]byte) []byte {
	if p.PKSigner == nil {
		p.PKSigner, _ = cr.LoadPrivateKey(p.myPrivKey)
	}

	sig, _ := p.PKSigner.Sign(hash[:])
	return sig
}
