package fcm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTopicHandle_1(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(topicHandle))
	defer srv.Close()

	c := NewFcmClient("key")
	c.FCMServerURL = srv.URL

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	c.NewFcmMsgTo("/topics/topicName", data)

	res, err := c.Send()
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}
}

func TestTopicHandle_2(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(topicHandle))
	defer srv.Close()

	c := NewFcmClient("key")
	c.FCMServerURL = srv.URL

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	c.NewFcmTopicMsg("/topics/topicName", data)

	res, err := c.Send()
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}
}

func TestTopicHandle_3(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(topicHandle))
	defer srv.Close()

	c := NewFcmClient("key")
	c.FCMServerURL = srv.URL

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	data2 := map[string]string{
		"msg": "Hello bits",
	}

	c.NewFcmTopicMsg("/topics/topicName", data)

	c.SetMsgData(data2)
	res, err := c.Send()
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}
}

func TestRegIdHandle_1(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(regIdHandle))
	defer srv.Close()

	c := NewFcmClient("key")
	c.FCMServerURL = srv.URL

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	ids := []string{
		"token0",
		"token1",
		"token2",
	}

	c.NewFcmRegIdsMsg(ids, data)

	res, err := c.Send()
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}

	if res.Success != 2 || res.Fail != 1 {
		t.Error("Parsing Success or Fail error")
	}
}

func TestRegIdHandle_2(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(regIdHandle))
	defer srv.Close()

	c := NewFcmClient("key")
	c.FCMServerURL = srv.URL

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	ids := []string{
		"token0",
	}

	xds := []string{
		"token1",
		"token2",
	}

	c.newDevicesList(ids)

	c.SetMsgData(data)

	c.AppendDevices(xds)

	res, err := c.Send()
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}

	if res.Success != 2 || res.Fail != 1 {
		t.Error("Parsing Success or Fail error")
	}
}

func TestSendWithContext(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(regIdHandle))
	defer srv.Close()

	c := NewFcmClient("key")
	c.FCMServerURL = srv.URL

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	c.NewFcmMsgTo("/topics/topicName", data)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res, err := c.SendWithContext(ctx)
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}

	// After the cancellation the request is expected to fail.
	cancel()

	_, err = c.SendWithContext(ctx)
	if err == nil {
		t.Errorf("expected context error, got %v", err)
	}

}

func topicHandle(w http.ResponseWriter, r *http.Request) {
	result := `{"message_id":6985435902064854329}`
	fmt.Fprintln(w, result)
}

func regIdHandle(w http.ResponseWriter, r *http.Request) {
	result := `{"multicast_id":1003859738309903334,"success":2,"failure":1,"canonical_ids":0,"results":[{"message_id":"0:1448128667408487%ecaaa23db3fd7efd"},{"message_id":"0:1468135657607438%ecafacddf9ff8ead"},{"error":"InvalidRegistration"}]}`
	fmt.Fprintln(w, result)
}

func testSend(t *testing.T, c *FcmClient) {
	res, err := c.Send()
	checkErrors(t, c, res, err)
}

func testSendContext(t *testing.T, c *FcmClient) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	res, err := c.SendWithContext(ctx)
	checkErrors(t, c, res, err)
}

func checkErrors(t *testing.T, c *FcmClient, res *FcmResponseStatus, err error) {
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}
	if c.Message.DirectBootOk != true {
		t.Error("Failed to set DirectBootOk to true")
	}
}

func TestSendDirectBoot(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(topicHandle))
	defer srv.Close()

	c := NewFcmClient("key")
	c.FCMServerURL = srv.URL

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	data2 := map[string]string{
		"msg": "Hello bits",
	}

	ids := []string{
		"token0",
		"token1",
		"token2",
	}

	c.NewFcmTopicMsg("/topics/topicName", data)
	testSend(t, c)

	c.SetMsgData(data2)
	testSend(t, c)

	c.NewFcmRegIdsMsg(ids, data)
	testSend(t, c)

	c.NewFcmTopicMsg("/topics/topicName", data)
	testSendContext(t, c)

	c.SetMsgData(data2)
	testSendContext(t, c)

	c.NewFcmRegIdsMsg(ids, data)
	testSendContext(t, c)
}
