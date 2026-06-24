# 🔧 Panduan & Prompt Refactor (Svelte + Go)

Dokumen ini berisi **prinsip, checklist, dan prompt siap pakai** untuk melakukan refactor kode dengan fokus pada:

- ✅ **Clean Code** — mudah dibaca, dirawat, dan dikembangkan.
- ✅ **Menyederhanakan logika** yang terlalu berbelit-belit (over-engineered / deeply nested).
- ✅ **Efisiensi** — performa eksekusi lebih cepat.
- ✅ **Hemat Resource** — minim pemakaian RAM, disk, CPU, dan I/O.

> **Stack:** Frontend = **Svelte / SvelteKit (TypeScript)** · Backend = **Go (Golang)**

---

## 📌 Aturan Emas Refactor

1. **Jangan ubah perilaku (behavior).** Refactor = memperbaiki struktur tanpa mengubah output/fungsionalitas.
2. **Satu perubahan, satu tujuan.** Refactor bertahap, bukan menulis ulang semuanya sekaligus.
3. **Ukur sebelum & sesudah.** Jangan mengoptimasi tanpa data (profiling/benchmark).
4. **Test dulu kalau memungkinkan.** Pastikan ada safety net sebelum mengubah.
5. **Hindari premature optimization.** Utamakan keterbacaan; optimasi resource pada bagian yang terbukti jadi bottleneck.

---

## 🧭 Prinsip Umum Clean Code

| Prinsip | Penjelasan |
|---|---|
| **DRY** | Don't Repeat Yourself — hilangkan duplikasi logika. |
| **KISS** | Keep It Simple — pilih solusi paling sederhana yang bekerja. |
| **YAGNI** | You Aren't Gonna Need It — jangan buat abstraksi untuk kebutuhan yang belum ada. |
| **SRP** | Single Responsibility — satu fungsi/komponen, satu tanggung jawab. |
| **Early Return** | Kurangi nesting dengan guard clause / return lebih awal. |
| **Naming** | Nama variabel/fungsi deskriptif & konsisten. |

---

## 🎯 Menyederhanakan Logika Berbelit-belit

Tanda logika perlu disederhanakan (**code smell**):

- Nesting `if/else` lebih dari 2–3 level.
- Fungsi panjang (> 50 baris) yang melakukan banyak hal.
- Banyak flag boolean untuk mengatur alur.
- Kondisi kompleks yang sulit dibaca (`if a && b || c && !d`).
- Duplikasi blok kode di banyak tempat.

**Teknik perbaikan:**

- **Guard clause / early return** untuk menghapus `else` bersarang.
- **Extract function** — pecah fungsi besar jadi fungsi kecil bernama jelas.
- **Lookup table / map** menggantikan rantai `if-else` atau `switch` panjang.
- **Pisahkan logika dari I/O** (pure function lebih mudah dites & dioptimasi).

```text
SEBELUM (berbelit):
if (user != null) {
  if (user.active) {
    if (user.hasAccess) {
      doSomething()
    }
  }
}

SESUDAH (early return):
if (user == null) return
if (!user.active) return
if (!user.hasAccess) return
doSomething()
```

---

## 🎨 Checklist Refactor Frontend (Svelte / SvelteKit)

### Struktur & Clean Code
- [ ] Pecah komponen besar → komponen kecil yang reusable.
- [ ] Pisahkan logika bisnis dari UI (pindah ke `stores`, `utils`, atau modul terpisah).
- [ ] Hilangkan duplikasi markup & logic.
- [ ] Gunakan TypeScript dengan tipe ketat (hindari `any`).
- [ ] Konsisten dalam penamaan file, props, dan event.

### Efisiensi & Hemat Resource
- [ ] **Reactivity tepat sasaran** — gunakan runes (`$state`, `$derived`, `$effect`) atau reactive statement hanya seperlunya agar tidak ada render ulang berlebih.
- [ ] **Hindari komputasi berat di reactive block** — gunakan `$derived`/memoization untuk hasil yang dipakai ulang.
- [ ] **Lazy load** komponen & route berat (`import()` dinamis) → hemat bundle awal & RAM.
- [ ] **Virtual list** untuk daftar panjang (jangan render ribuan node DOM sekaligus).
- [ ] **Bersihkan side effect** — selalu `return` cleanup di `$effect`/`onMount` (hindari memory leak: listener, interval, subscription).
- [ ] **Optimasi aset** — kompres gambar, gunakan format modern (WebP/AVIF), `loading="lazy"`.
- [ ] **Kurangi dependency** pihak ketiga yang besar → bundle lebih kecil, disk & memori lebih hemat.
- [ ] **Debounce/throttle** event yang sering terpicu (scroll, input, resize).
- [ ] Hindari menyimpan data besar yang tidak perlu di store global.

---

## ⚙️ Checklist Refactor Backend (Go)

### Struktur & Clean Code
- [ ] Terapkan pemisahan layer (handler → service → repository).
- [ ] Fungsi pendek dengan satu tanggung jawab (SRP).
- [ ] Error handling eksplisit & konsisten (`if err != nil`, wrap dengan `fmt.Errorf("...: %w", err)`).
- [ ] Gunakan `interface` untuk dependency agar mudah dites & di-mock.
- [ ] Hindari package global state; gunakan dependency injection sederhana.

### Efisiensi & Hemat Resource (RAM, CPU, Disk, I/O)
- [ ] **Hindari alokasi memori tak perlu:**
  - Preallocate slice/map dengan kapasitas: `make([]T, 0, n)` / `make(map[K]V, n)`.
  - Gunakan `strings.Builder` untuk concat string (hindari `+=` di loop).
  - Pakai pointer untuk struct besar agar tidak menyalin nilai.
- [ ] **Kelola goroutine dengan benar** — batasi jumlah (worker pool), hindari goroutine leak, gunakan `context` untuk cancellation.
- [ ] **Tutup resource** dengan `defer` (file, rows, response body) agar tidak bocor handle/disk.
- [ ] **Streaming, bukan load semua ke memori** — proses file/data besar via `io.Reader`/streaming, bukan `ReadAll` ke RAM.
- [ ] **Optimasi database:**
  - Hindari N+1 query.
  - `SELECT` kolom yang dibutuhkan saja.
  - Gunakan index & connection pooling (`SetMaxOpenConns`, `SetMaxIdleConns`).
  - Gunakan batch operation untuk insert/update massal.
- [ ] **Caching** hasil komputasi/query yang mahal & sering dipakai (dengan TTL agar memori terkendali).
- [ ] **Gunakan `sync.Pool`** untuk objek yang sering dibuat-dibuang (mengurangi tekanan GC).
- [ ] **Reuse buffer** dan hindari konversi `[]byte`↔`string` berulang yang memicu alokasi.
- [ ] **Profiling** dengan `pprof` dan `go test -bench -benchmem` untuk membuktikan perbaikan.

---

## 🤖 Prompt Siap Pakai

### Prompt Refactor Frontend (Svelte)

```
Kamu adalah senior frontend engineer ahli Svelte/SvelteKit + TypeScript.
Refactor kode berikut TANPA mengubah perilaku yang terlihat user.

Tujuan:
1. Clean code: pisahkan logika dari UI, pecah komponen besar, hilangkan duplikasi.
2. Sederhanakan logika berbelit (early return, hindari nesting dalam).
3. Efisiensi: optimalkan reactivity, hindari render ulang & komputasi berlebih.
4. Hemat resource: lazy load, virtual list untuk data besar, bersihkan side effect
   (cegah memory leak), kurangi dependency berat, optimasi aset.
5. Typing ketat (hindari `any`).

Aturan:
- Jaga API props & fungsionalitas. Jelaskan setiap perubahan penting.
- Berikan kode lengkap + ringkasan code smell yang ditemukan.

Kode:
[TEMPEL KODE SVELTE DI SINI]
```

### Prompt Refactor Backend (Go)

```
Kamu adalah senior backend engineer ahli Go (Golang).
Refactor kode berikut TANPA mengubah kontrak API / perilaku.

Tujuan:
1. Clean code: pisahkan layer (handler/service/repository), fungsi pendek (SRP),
   error handling konsisten (wrap dengan %w), gunakan interface untuk DI.
2. Sederhanakan logika berbelit (guard clause, lookup map ganti if-else panjang).
3. Efisiensi & hemat resource (RAM/CPU/Disk/I/O):
   - Preallocate slice/map, pakai strings.Builder, pointer untuk struct besar.
   - Tutup resource dengan defer, cegah goroutine/memory leak, pakai context.
   - Streaming untuk data besar (hindari ReadAll ke RAM).
   - Optimasi query DB (hindari N+1, index, connection pool, batch).
   - Pertimbangkan sync.Pool & caching untuk objek/komputasi mahal.

Aturan:
- Jaga kompatibilitas pemanggil. Jelaskan keputusan refactor.
- Berikan kode lengkap + ringkasan code smell + saran benchmark (pprof/benchmem).

Kode:
[TEMPEL KODE GO DI SINI]
```

---

## 📊 Verifikasi Setelah Refactor

| Aspek | Frontend (Svelte) | Backend (Go) |
|---|---|---|
| **Fungsionalitas** | Output & interaksi UI tetap sama | Response API & status code tetap sama |
| **Performa** | Lighthouse, ukuran bundle, jumlah render | `go test -bench`, `pprof`, latensi |
| **Memori** | DevTools Memory profiler, cek leak | `-benchmem`, heap profile, cek leak |
| **Disk/Bundle** | Ukuran build/bundle berkurang | Ukuran binary, penggunaan disk I/O |
| **Test** | Unit/E2E hijau | `go test ./...` hijau |

---

## ✅ Definition of Done

- [ ] Perilaku tidak berubah (terverifikasi test/manual).
- [ ] Logika lebih sederhana & mudah dibaca.
- [ ] Tidak ada duplikasi & dead code.
- [ ] Resource (RAM/CPU/disk/bundle) terbukti sama atau lebih hemat (ada data).
- [ ] Tidak ada resource/memory leak.
- [ ] Perubahan terdokumentasi (alasan & breaking change jika ada).