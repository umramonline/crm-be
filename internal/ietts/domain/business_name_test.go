package domain

import "testing"

func TestSplitBusinessNameSplitsFirstWordToAd(t *testing.T) {
	ad, soyad := SplitBusinessName("CENGİZ ÇITANAK")
	if ad != "CENGİZ" {
		t.Fatalf("expected ad CENGİZ, got %q", ad)
	}
	if soyad != "ÇITANAK" {
		t.Fatalf("expected soyad ÇITANAK, got %q", soyad)
	}
}

func TestSplitBusinessNameJoinsRemainingWordsToSoyad(t *testing.T) {
	ad, soyad := SplitBusinessName("AHMET MEHMET YILMAZ")
	if ad != "AHMET" {
		t.Fatalf("expected ad AHMET, got %q", ad)
	}
	if soyad != "MEHMET YILMAZ" {
		t.Fatalf("expected soyad MEHMET YILMAZ, got %q", soyad)
	}
}

func TestSplitBusinessNameLeavesSoyadEmptyForSingleWord(t *testing.T) {
	ad, soyad := SplitBusinessName("BURAK")
	if ad != "BURAK" {
		t.Fatalf("expected ad BURAK, got %q", ad)
	}
	if soyad != "" {
		t.Fatalf("expected empty soyad, got %q", soyad)
	}
}

func TestSplitBusinessNameReturnsEmptyForBlankValue(t *testing.T) {
	ad, soyad := SplitBusinessName("   ")
	if ad != "" || soyad != "" {
		t.Fatalf("expected empty ad and soyad, got %q %q", ad, soyad)
	}
}
