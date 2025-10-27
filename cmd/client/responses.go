package main

// func (c *Client) Client2GatewayResponse(target string, res models.Response) error {
// 	var topic string
// 	if target == "" {
// 		topic = pubsub.GetMQTTGatewayResponseBroadcast("")
// 	} else {
// 		topic = pubsub.GetMQTTGatewayResponseTopic(target, "")
// 	}
// 	log.Println(topic)
// 	pubsub.PublishMQTT(
// 		c.Ctx,
// 		c.MQTTClient,
// 		pubsub.MQTTPubOpts{
// 			Topic: topic,
// 			QoS:   1,
// 		},
// 		models.ContentJSON,
// 		res,
// 	)
// 	return nil
// }
