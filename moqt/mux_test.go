package moqt

import (
	"errors"
	"sync"
	"testing"
)

// MockTrackHandler はTrackHandlerインターフェースのモック実装
type MockTrackHandler struct {
	ServeTrackFunc         func(w TrackWriter, config SubscribeConfig)
	GetInfoFunc            func(path TrackPath) (Info, error)
	ServeAnnouncementsFunc func(w AnnouncementWriter)
	HandlerCalled          bool
	Path                   TrackPath
}

func (m *MockTrackHandler) ServeTrack(w TrackWriter, config SubscribeConfig) {
	m.HandlerCalled = true
	if m.ServeTrackFunc != nil {
		m.ServeTrackFunc(w, config)
	}
}

func (m *MockTrackHandler) GetInfo(path TrackPath) (Info, error) {
	if m.GetInfoFunc != nil {
		return m.GetInfoFunc(path)
	}
	return Info{Path: path}, nil
}

func (m *MockTrackHandler) ServeAnnouncements(w AnnouncementWriter) {
	if m.ServeAnnouncementsFunc != nil {
		m.ServeAnnouncementsFunc(w)
	}
}

// MockTrackWriter はTrackWriterインターフェースのモック実装
type MockTrackWriter struct {
	PathValue      TrackPath
	TrackDataValue []byte
	WrittenData    []byte
}

func (m *MockTrackWriter) TrackPath() TrackPath {
	return m.PathValue
}

func (m *MockTrackWriter) Write(data []byte) (int, error) {
	m.WrittenData = append(m.WrittenData, data...)
	return len(data), nil
}

// MockAnnouncementWriter はAnnouncementWriterインターフェースのモック実装
type MockAnnouncementWriter struct {
	ConfigValue      AnnounceConfig
	AnnouncedTracks  []TrackPath
	AnnouncementData []byte
	Notifications    int
}

func (m *MockAnnouncementWriter) AnnounceConfig() AnnounceConfig {
	return m.ConfigValue
}

func (m *MockAnnouncementWriter) WriteAnnouncement(path TrackPath, data []byte) (int, error) {
	m.AnnouncedTracks = append(m.AnnouncedTracks, path)
	m.AnnouncementData = append(m.AnnouncementData, data...)
	m.Notifications++
	return len(data), nil
}

// TestTrackMuxBasicRouting は基本的なルーティング機能をテスト
func TestTrackMuxBasicRouting(t *testing.T) {
	mux := NewTrackMux()

	// ハンドラーの作成とパスへの登録
	audioHandler := &MockTrackHandler{Path: "/tracks/audio"}
	videoHandler := &MockTrackHandler{Path: "/tracks/video"}

	mux.Handle("/tracks/audio", audioHandler)
	mux.Handle("/tracks/video", videoHandler)

	// オーディオトラックへのリクエストを検証
	audioWriter := &MockTrackWriter{PathValue: "/tracks/audio"}
	mux.ServeTrack(audioWriter, SubscribeConfig{})

	if !audioHandler.HandlerCalled {
		t.Errorf("Audio handler was not called for /tracks/audio path")
	}

	// ビデオトラックへのリクエストを検証
	videoWriter := &MockTrackWriter{PathValue: "/tracks/video"}
	mux.ServeTrack(videoWriter, SubscribeConfig{})

	if !videoHandler.HandlerCalled {
		t.Errorf("Video handler was not called for /tracks/video path")
	}

	// 存在しないパスへのリクエストを検証
	unknownWriter := &MockTrackWriter{PathValue: "/tracks/unknown"}
	notFoundCalled := false
	NotFoundHandler = &MockTrackHandler{
		ServeTrackFunc: func(w TrackWriter, config SubscribeConfig) {
			notFoundCalled = true
		},
	}
	mux.ServeTrack(unknownWriter, SubscribeConfig{})

	if !notFoundCalled {
		t.Errorf("NotFoundHandler was not called for non-existent path")
	}
}

// TestGetInfo はGetInfo機能をテスト
func TestGetInfo(t *testing.T) {
	mux := NewTrackMux()

	expectedInfo := Info{
		Path: "/tracks/audio",
		Name: "Audio Track",
	}

	audioHandler := &MockTrackHandler{
		GetInfoFunc: func(path TrackPath) (Info, error) {
			return expectedInfo, nil
		},
	}

	mux.Handle("/tracks/audio", audioHandler)

	// 存在するトラックの情報取得をテスト
	info, err := mux.GetInfo("/tracks/audio")
	if err != nil {
		t.Errorf("GetInfo returned error: %v", err)
	}

	if info.Name != expectedInfo.Name {
		t.Errorf("Expected info name %s, got %s", expectedInfo.Name, info.Name)
	}

	// 存在しないトラックをテスト
	_, err = mux.GetInfo("/tracks/unknown")
	if err == nil || !errors.Is(err, ErrTrackDoesNotExist) {
		t.Errorf("Expected ErrTrackDoesNotExist for non-existent track, got: %v", err)
	}
}

// TestAnnouncements はアナウンスメントシステムをテスト
func TestAnnouncements(t *testing.T) {
	mux := NewTrackMux()

	// まずアナウンスメントの購読者を登録
	audioAnnouncer := &MockAnnouncementWriter{
		ConfigValue: AnnounceConfig{
			TrackPattern: "/tracks/audio/*", // audio配下の単一セグメント
		},
	}

	videoAnnouncer := &MockAnnouncementWriter{
		ConfigValue: AnnounceConfig{
			TrackPattern: "/tracks/video/**", // video配下の複数セグメント
		},
	}

	allTracksAnnouncer := &MockAnnouncementWriter{
		ConfigValue: AnnounceConfig{
			TrackPattern: "/**", // すべてのトラック
		},
	}

	mux.ServeAnnouncements(audioAnnouncer)
	mux.ServeAnnouncements(videoAnnouncer)
	mux.ServeAnnouncements(allTracksAnnouncer)

	// ハンドラーの登録でアナウンスメントが発生するか検証
	audioHandler := &MockTrackHandler{
		ServeAnnouncementsFunc: func(w AnnouncementWriter) {
			w.WriteAnnouncement("/tracks/audio/main", []byte("audio announcement"))
		},
	}

	videoHandler := &MockTrackHandler{
		ServeAnnouncementsFunc: func(w AnnouncementWriter) {
			w.WriteAnnouncement("/tracks/video/main", []byte("video announcement"))
		},
	}

	nestedVideoHandler := &MockTrackHandler{
		ServeAnnouncementsFunc: func(w AnnouncementWriter) {
			w.WriteAnnouncement("/tracks/video/streams/hd", []byte("nested video announcement"))
		},
	}

	otherHandler := &MockTrackHandler{
		ServeAnnouncementsFunc: func(w AnnouncementWriter) {
			w.WriteAnnouncement("/other/track", []byte("other announcement"))
		},
	}

	mux.Handle("/tracks/audio/main", audioHandler)
	mux.Handle("/tracks/video/main", videoHandler)
	mux.Handle("/tracks/video/streams/hd", nestedVideoHandler)
	mux.Handle("/other/track", otherHandler)

	// 各アナウンサーが正しいパターンのトラックを受信したか検証
	if audioAnnouncer.Notifications != 1 {
		t.Errorf("Audio announcer should receive 1 notification, got %d", audioAnnouncer.Notifications)
	}

	if videoAnnouncer.Notifications != 2 {
		t.Errorf("Video announcer should receive 2 notifications, got %d", videoAnnouncer.Notifications)
	}

	if allTracksAnnouncer.Notifications != 4 {
		t.Errorf("All tracks announcer should receive 4 notifications, got %d", allTracksAnnouncer.Notifications)
	}
}

// TestHandlerOverwrite はハンドラーの上書きをテスト
func TestHandlerOverwrite(t *testing.T) {
	mux := NewTrackMux()

	handler1 := &MockTrackHandler{}
	handler2 := &MockTrackHandler{}

	mux.Handle("/tracks/test", handler1)

	// 同じパスに別のハンドラを登録する（警告が記録されるはず）
	mux.Handle("/tracks/test", handler2)

	// ハンドラ2が使用されることを確認
	writer := &MockTrackWriter{PathValue: "/tracks/test"}
	mux.ServeTrack(writer, SubscribeConfig{})

	if !handler2.HandlerCalled {
		t.Errorf("Handler2 should be called after overwriting handler1")
	}
}

// TestConcurrentAccess は同時アクセスの安全性をテスト
func TestConcurrentAccess(t *testing.T) {
	mux := NewTrackMux()

	// 複数のハンドラをセットアップ
	for i := 0; i < 10; i++ {
		path := TrackPath("/tracks/path" + string(rune(i+'0')))
		handler := &MockTrackHandler{Path: path}
		mux.Handle(path, handler)
	}

	var wg sync.WaitGroup

	// 10のGoroutineで同時に読み取り
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := TrackPath("/tracks/path" + string(rune(idx+'0')))
			writer := &MockTrackWriter{PathValue: path}
			mux.ServeTrack(writer, SubscribeConfig{})
			mux.GetInfo(path)
		}(i)
	}

	// さらに5つのGoroutineで書き込み
	for i := 10; i < 15; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := TrackPath("/tracks/path" + string(rune(idx+'0')))
			handler := &MockTrackHandler{Path: path}
			mux.Handle(path, handler)
		}(i)
	}

	wg.Wait()
	// ここまで来れれば、デッドロックなしで同時アクセスが機能している
}

// TestDefaultMux はグローバルなDefaultMuxの機能をテスト
func TestDefaultMux(t *testing.T) {
	// リセット用の新しいMuxを一時的に保存
	origDefaultMux := DefaultMux
	defer func() {
		DefaultMux = origDefaultMux
	}()

	DefaultMux = NewTrackMux()

	// デフォルトのMuxにハンドラを登録
	handler := &MockTrackHandler{}
	Handle("/default/test", handler)

	// ハンドラが正しく呼ばれるか確認
	writer := &MockTrackWriter{PathValue: "/default/test"}
	ServeTrack(writer, SubscribeConfig{})

	if !handler.HandlerCalled {
		t.Errorf("Handler was not called via DefaultMux")
	}

	// GetInfoもテスト
	info, err := GetInfo("/default/test")
	if err != nil {
		t.Errorf("GetInfo via DefaultMux returned error: %v", err)
	}

	if info.Path != "/default/test" {
		t.Errorf("Expected path /default/test, got %s", info.Path)
	}
}

// TestWildcardRouting はワイルドカードのパスマッチングをテスト
func TestWildcardRouting(t *testing.T) {
	mux := NewTrackMux()

	singleHandler := &MockTrackHandler{}
	doubleHandler := &MockTrackHandler{}

	// シングルワイルドカード（*）を購読するアナウンサー
	singleWildcardAnnouncer := &MockAnnouncementWriter{
		ConfigValue: AnnounceConfig{
			TrackPattern: "/wildcard/*",
		},
	}

	// ダブルワイルドカード（**）を購読するアナウンサー
	doubleWildcardAnnouncer := &MockAnnouncementWriter{
		ConfigValue: AnnounceConfig{
			TrackPattern: "/deep/**",
		},
	}

	mux.ServeAnnouncements(singleWildcardAnnouncer)
	mux.ServeAnnouncements(doubleWildcardAnnouncer)

	// 各パターンに対応するハンドラを登録
	mux.Handle("/wildcard/one", singleHandler)
	mux.Handle("/deep/one/two/three", doubleHandler)

	// シングルワイルドカード（*）のテスト
	if singleWildcardAnnouncer.Notifications != 1 {
		t.Errorf("Single wildcard announcer should receive 1 notification, got %d", singleWildcardAnnouncer.Notifications)
	}

	// ダブルワイルドカード（**）のテスト
	if doubleWildcardAnnouncer.Notifications != 1 {
		t.Errorf("Double wildcard announcer should receive 1 notification, got %d", doubleWildcardAnnouncer.Notifications)
	}
}

// BenchmarkPathMatching はパスマッチングのパフォーマンスをベンチマーク
func BenchmarkPathMatching(b *testing.B) {
	mux := NewTrackMux()

	// 多数のハンドラを登録
	for i := 0; i < 100; i++ {
		for j := 0; j < 10; j++ {
			path := TrackPath("/section" + string(rune(i+'0')) + "/subsection" + string(rune(j+'0')))
			mux.Handle(path, &MockTrackHandler{})
		}
	}

	// 深い階層のパスをテスト
	writer := &MockTrackWriter{PathValue: "/section5/subsection7"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mux.ServeTrack(writer, SubscribeConfig{})
	}
}
