package crypto

//
//import (
//	"fmt"
//
//	"utils/cgo/crypto"
//)
//
//const (
//	ESEA_NO_ENCRYPT          int = 0 //!<直接拷贝，不加密
//	ESEA_OI_SYMMETRY_ENCRYPT int = 1 //!<使用OI库的对称加密
//)
//
//func EncryptData(algorithm int, key []byte, in []byte) (out []byte, err error) {
//	if len(key) <= 0 {
//		return nil, fmt.Errorf("null key")
//	}
//
//	switch algorithm {
//	case ESEA_NO_ENCRYPT:
//		out = make([]byte, len(in))
//		copy(out, in)
//		return out, nil
//	case ESEA_OI_SYMMETRY_ENCRYPT:
//		if out, err := crypto.Oi_symmetry_encrypt2(in, key); err != nil {
//			return nil, err
//		} else {
//			return out, nil
//		}
//	default:
//		return nil, fmt.Errorf("invalid algorithm")
//	}
//	return nil, fmt.Errorf("invalid algorithm")
//}
//
//// 该加密函数，在key不变的情况下，OI库的TEA算法每次得到的密文是相同的
//func EncryptDataRegular(algorithm int, key []byte, in []byte) (out []byte, err error) {
//	if len(key) <= 0 {
//		return nil, fmt.Errorf("null key")
//	}
//
//	switch algorithm {
//	case ESEA_NO_ENCRYPT:
//		out = make([]byte, len(in))
//		copy(out, in)
//		return out, nil
//	case ESEA_OI_SYMMETRY_ENCRYPT:
//		if out, err := crypto.Oi_symmetry_encrypt2_regular(in, key); err != nil {
//			return nil, err
//		} else {
//			return out, nil
//		}
//	default:
//		return nil, fmt.Errorf("invalid algorithm")
//	}
//	return nil, fmt.Errorf("invalid algorithm")
//}
//
//func DecryptData(algorithm int, key []byte, in []byte) (out []byte, err error) {
//	if len(key) <= 0 {
//		return nil, fmt.Errorf("null key")
//	}
//
//	switch algorithm {
//	case ESEA_NO_ENCRYPT:
//		out = make([]byte, len(in))
//		copy(out, in)
//		return out, nil
//	case ESEA_OI_SYMMETRY_ENCRYPT:
//		if out, err := crypto.Oi_symmetry_decrypt2(in, key); err != nil {
//			return nil, err
//		} else {
//			return out, nil
//		}
//	default:
//		return nil, fmt.Errorf("invalid algorithm")
//	}
//	return nil, fmt.Errorf("invalid algorithm")
//}
