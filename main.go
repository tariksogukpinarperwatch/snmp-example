package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/hallidave/mibtool/smi"
)

func main() {
	// MIB dizinini belirtin
	mibDir := "C:/Users/Tarık/Desktop/go-snmp-example/standard_mibs"

	mib := smi.NewMIB(mibDir)
	mib.Debug = true
	// Yüklenen MIB dosyalarını ekleyin
	err := mib.LoadModules("IF-MIB")
	if err != nil {
		log.Fatalf("MIB yükleme hatası: %v", err)
	}

	fmt.Println("Yüklenen MIB Modülleri:")
	for modName := range mib.Modules {
		fmt.Println(" -", modName)
	}

	// SNMP Client yapılandırması
	go_service := &gosnmp.GoSNMP{
		Target:    "192.168.20.201", // Hedef cihazın IP adresi
		Port:      161,              // SNMP varsayılan port
		Version:   gosnmp.Version2c, // SNMP sürümü
		Community: "public",         // Topluluk adı
		Timeout:   5 * time.Second,  // Zaman aşımı
		Retries:   3,                // Yeniden deneme sayısı
	}

	// SNMP bağlantısı
	err = go_service.Connect()
	if err != nil {
		log.Fatalf("SNMP bağlantısı başarısız: %v", err)
	}
	defer go_service.Conn.Close()

	// Walk ile tüm OID’leri keşfetme
	fmt.Println("Tüm OID'leri keşfediyoruz...")
	var oids []string
	err = go_service.Walk(".1.3.6.1", func(pdu gosnmp.SnmpPDU) error {
		if pdu.Value == nil {
			fmt.Printf("OID: %s returned nil, skipping...\n", pdu.Name)
			return nil
		}
		oids = append(oids, pdu.Name)
		oidName := resolveOID(mib, pdu.Name)
		switch pdu.Type {
		case gosnmp.OctetString:
			fmt.Printf("OID: %s (%s), Value: %s\n", pdu.Name, oidName, string(pdu.Value.([]byte)))
		default:
			fmt.Printf("OID: %s (%s), Value: %v\n", pdu.Name, oidName, pdu.Value)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("SNMP Walk işlemi başarısız: %v", err)
	}

	// Get metodu ile OID'leri test etme
	fmt.Println("\nGet metodu ile OID'leri test ediyoruz...")
	for _, oid := range oids {
		result, err := go_service.Get([]string{oid})
		if err != nil {
			fmt.Printf("OID: %s, Get başarısız: %v\n", oid, err)
			continue
		}

		for _, variable := range result.Variables {
			oidName := resolveOID(mib, variable.Name)
			switch variable.Type {
			case gosnmp.OctetString:
				fmt.Printf("OID: %s (%s), Value: %s (Get Test)\n", variable.Name, oidName, string(variable.Value.([]byte)))
			default:
				fmt.Printf("OID: %s (%s), Value: %v (Get Test)\n", variable.Name, oidName, variable.Value)
			}
		}
	}
	fmt.Printf("MIB yüklendi: %d modül\n", len(mib.Modules))
}

// OID anlamlandırma
func resolveOID(mib *smi.MIB, oid string) string {
	name, err := mib.OID(oid)
	if err != nil {
		return "Bilinmeyen OID"
	}
	return name.String()
}
