package tlsgen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// helper 写文件
func writePem(filename string, block *pem.Block) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建文件 %s 失败: %v", filename, err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			return
		}
	}(f)
	err = pem.Encode(f, block)
	if err != nil {
		return fmt.Errorf("写入文件 %s 失败: %v", filename, err)
	}
	return nil
}

// GenerateCA 生成 CA 根证书及私钥，写入 dir/ca.crt 和 dir/ca.key
func GenerateCA(dir string, validDays int, overwrite bool) error {
	caKeyPath := filepath.Join(dir, "ca.key")
	caCrtPath := filepath.Join(dir, "ca.crt")
	var crtIsExists bool
	var keyIsExists bool
	// 检测文件是否存在
	if _, err := os.Stat(caKeyPath); err == nil {
		crtIsExists = true
	}
	if _, err := os.Stat(caCrtPath); err == nil {
		keyIsExists = true
	}
	// 如果文件存在且 overwrite 为 false，则不生成新的证书
	if crtIsExists && keyIsExists && !overwrite {
		log.Println("CA 证书已存在，如要重新生成，请使用 -overwrite 选项")
		return nil
	}
	// 创建目录
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return fmt.Errorf("创建目录 %s 失败: %v", dir, err)
	}
	// 生成 CA 私钥和证书
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("生成 CA 密钥失败: %v", err)
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("生成 CA 序列号失败: %v", err)
	}
	caTemplate := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "GoRootCA",
			Organization: []string{"MyOrg"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, validDays),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("创建 CA 证书失败: %v", err)
	}
	err = writePem(dir+"/ca.key", &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caKey)})
	if err != nil {
		return fmt.Errorf("写入 CA 密钥失败: %v", err)
	}
	err = writePem(dir+"/ca.crt", &pem.Block{Type: "CERTIFICATE", Bytes: caDER})
	if err != nil {
		return fmt.Errorf("写入 CA 证书失败: %v", err)
	}
	log.Println("生成 CA 证书成功")
	return nil
}

// GenerateServerAndClientCerts 基于已有 CA 生成 server 和 client 证书
func GenerateServerAndClientCerts(dir string, validDays int, caCertPath, caKeyPath string) error {
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return fmt.Errorf("创建目录 %s 失败: %v", dir, err)
	}

	// 读取 CA 证书
	caCertPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		return fmt.Errorf("读取 CA 证书失败: %v", err)
	}
	block, _ := pem.Decode(caCertPEM)
	if block == nil {
		return fmt.Errorf("解析 CA 证书失败: %v", err)
	}
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("解析 CA 证书失败: %v", err)
	}

	// 读取 CA 私钥
	caKeyPEM, err := os.ReadFile(caKeyPath)
	if err != nil {
		return fmt.Errorf("读取 CA 私钥失败: %v", err)
	}
	block, _ = pem.Decode(caKeyPEM)
	if block == nil {
		return errors.New("解析 CA 私钥失败")
	}
	caKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("解析 CA 私钥失败: %v", err)
	}

	// =============== 生成服务端证书 =================
	err = GenerateServer(dir, validDays, caCert, caKey)
	if err != nil {
		return fmt.Errorf("生成服务端证书失败: %v", err)
	}
	// =============== 生成客户端证书 =================
	err = GenerateClient(dir, validDays, caCert, caKey)
	if err != nil {
		return fmt.Errorf("生成客户端证书失败: %v", err)
	}

	log.Println("生成服务端和客户端证书成功")
	return nil
}

func GenerateServer(dir string, validDays int, caCert *x509.Certificate, caKey *rsa.PrivateKey) error {
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("生成服务端密钥失败: %v", err)
	}
	serialServer, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("生成服务端序列号失败: %v", err)
	}
	serverTemplate := &x509.Certificate{
		SerialNumber: serialServer,
		Subject: pkix.Name{
			CommonName:   "localhost",
			Organization: []string{"MyServer"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(0, 0, validDays),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:    []string{"localhost"},
	}

	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("创建服务端证书失败: %v", err)
	}
	err = writePem(dir+"/server.key", &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverKey)})
	if err != nil {
		return fmt.Errorf("写入服务端密钥失败: %v", err)
	}
	err = writePem(dir+"/server.crt", &pem.Block{Type: "CERTIFICATE", Bytes: serverDER})
	if err != nil {
		return fmt.Errorf("写入服务端证书失败: %v", err)
	}
	return nil
}

func GenerateClient(dir string, validDays int, caCert *x509.Certificate, caKey *rsa.PrivateKey) error {
	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("生成客户端密钥失败: %v", err)
	}
	serialClient, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("生成客户端序列号失败: %v", err)
	}
	clientTemplate := &x509.Certificate{
		SerialNumber: serialClient,
		Subject: pkix.Name{
			CommonName:   "client",
			Organization: []string{"MyClient"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(0, 0, validDays),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	clientDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, caCert, &clientKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("创建客户端证书失败: %v", err)
	}

	err = writePem(dir+"/client.key", &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientKey)})
	if err != nil {
		return fmt.Errorf("写入客户端密钥失败: %v", err)
	}
	err = writePem(dir+"/client.crt", &pem.Block{Type: "CERTIFICATE", Bytes: clientDER})
	if err != nil {
		return fmt.Errorf("写入客户端证书失败: %v", err)
	}
	return nil
}
