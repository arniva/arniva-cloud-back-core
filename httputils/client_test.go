package httputils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
)

type yankiIstek struct {
	No int `json:"no"`
}

type yankiCevap struct {
	Sunucu string `json:"sunucu"`
	No     int    `json:"no"`
	Baslik string `json:"baslik"`
}

// yankiSunucusu yalnızca kendi yolunu kabul eder; isteğin sıra numarasını
// (POST'ta gövdeden, GET'te sorgudan) ve X-No header'ını geri yankılar. Başka
// yola gelen isteğe 404 döner. Dönen değer sunucunun tam adresidir.
func yankiSunucusu(t *testing.T, ad, yol string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Her istek kendi bağlantısını açıyor; test tekrarlarında boşta bağlantı
		// birikip yerel portları tüketmesin.
		w.Header().Set("Connection", "close")
		if r.URL.Path != yol {
			http.Error(w, "yanlış yol: "+r.URL.Path, http.StatusNotFound)
			return
		}
		var ist yankiIstek
		if r.Method == http.MethodPost {
			if err := json.NewDecoder(r.Body).Decode(&ist); err != nil {
				http.Error(w, "gövde okunamadı: "+err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			ist.No, _ = strconv.Atoi(r.URL.Query().Get("no"))
		}
		json.NewEncoder(w).Encode(yankiCevap{Sunucu: ad, No: ist.No, Baslik: r.Header.Get("X-No")})
	}))
	t.Cleanup(srv.Close)
	return srv.URL + yol
}

// sendRequest fiber Agent'ını havuza bir kez bırakmalı. İki kez bırakılırsa
// havuzda aynı Agent iki kez durur; eşzamanlı iki istek aynı nesneyi alıp
// birbirinin adresini, header'larını ve gövdesini ezer (yanlış yere giden
// istek, başka isteğin cevabı, nil pointer panic).
func TestEsZamanliIsteklerKarismaz(t *testing.T) {
	sunucular := [2]struct{ ad, adres string }{
		{"A", yankiSunucusu(t, "A", "/a")},
		{"B", yankiSunucusu(t, "B", "/b")},
	}
	const goroutineSayisi, tur = 32, 25
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		hatalar []string
	)
	kaydet := func(s string) {
		mu.Lock()
		hatalar = append(hatalar, s)
		mu.Unlock()
	}
	basla := make(chan struct{})
	for g := 0; g < goroutineSayisi; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			<-basla
			// Art arda istekler: havuzdaki Agent'lar ancak ilk istekler
			// bittikten sonra yeniden dağıtılıyor.
			for i := 0; i < tur; i++ {
				no := g*tur + i
				func() {
					defer func() {
						if r := recover(); r != nil {
							kaydet(fmt.Sprintf("#%d panik: %v", no, r))
						}
					}()
					s := sunucular[no%2]
					opt := Option{Headers: map[string]string{"X-No": strconv.Itoa(no)}}
					var (
						res *HTTPResponse[yankiCevap]
						err error
					)
					if (no/2)%2 == 0 {
						res, err = Post[yankiCevap](s.adres, yankiIstek{No: no}, opt)
					} else {
						res, err = Get[yankiCevap](s.adres+"?no="+strconv.Itoa(no), opt)
					}
					switch {
					case err != nil:
						kaydet(fmt.Sprintf("#%d (%s): %v", no, s.ad, err))
					case res.Body.Sunucu != s.ad || res.Body.No != no || res.Body.Baslik != strconv.Itoa(no):
						kaydet(fmt.Sprintf("#%d (%s): başka isteğin cevabı geldi %+v", no, s.ad, res.Body))
					}
				}()
			}
		}(g)
	}
	close(basla)
	wg.Wait()
	if len(hatalar) > 0 {
		t.Errorf("%d istekten %d tanesi bozuldu:\n%s", goroutineSayisi*tur, len(hatalar),
			strings.Join(hatalar[:min(len(hatalar), 10)], "\n"))
	}
}
