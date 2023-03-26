/*
 TODO: packageの説明を書く
*/

package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	enc = binary.BigEndian
)

const (
	lenWidth = 8
)

type store struct {
	// ファイルへのポインタ
	*os.File
	// mutex
	mu sync.Mutex
	// bufioのWriter
	buf *bufio.Writer
	// logのサイズ
	size uint64
}

/*
新規にlogを作成するためのmethod
*/
func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	size := uint64(fi.Size())
	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f),
	}, nil
}

/*
 store構造体のメソッド。byte型でmessageを受け取り、保存する

 return:
  n uint64 レコードの長さ
  pos uint64 positionの略。レコードの位置を返す
  err error エラー
*/
func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	// Storeのlock
	s.mu.Lock()
	defer s.mu.Unlock()

	// 書き込み前のサイズをpositionとして覚えておき、呼び出し元に返す。
	pos = s.size
	// レコードの長さをまず書く。レコードを読み出すときに何バイト読み出せばいいかわかるように。失敗したらerrorを返す
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}
	// 書き込みされたバイト数が返される
	w, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}
	// header分を足し合わせて、レコードサイズとする
	w += lenWidth
	s.size += uint64(w)
	return uint64(w), pos, nil
}

/*
 store構造体のメソッド。positionを指定して、読み取りを行う
*/
func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}
	// レコードのheaderを読んで、レコードのサイズを取得する
	size := make([]byte, lenWidth)
	// 直前に取得したレコードのサイズと、メソッドの引数として渡されているポジションを渡してreadする。
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}
	// レコードサイズ分の容量を確保する
	b := make([]byte, enc.Uint64(size))
	// 確保したバッファ（b）に、読み込んだ内容を格納。
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}
	return b, nil
}

/*
 store構造体のメソッド。
*/
func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

/*
 ファイルをクローズする前にバッファされたデータを永続化する
*/
func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.buf.Flush()
	if err != nil {
		return err
	}
	return s.File.Close()
}
