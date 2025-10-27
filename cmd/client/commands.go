package main

// ========================
// Send
// ========================

// func (c *Client) SendGatewayCommand(gatewayID string, command models.Command) error {
// 	topic := pubsub.GetMQTTGatewayResponseTopic(gatewayID, "")

// 	err := pubsub.PublishMQTT(
// 		c.Ctx,
// 		c.MQTTClient,
// 		pubsub.MQTTPubOpts{
// 			Topic: topic,
// 			QoS:   1,
// 		},
// 		models.ContentJSON,
// 		command,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("failed to send command to topic %s: %w", topic, err)
// 	}
// 	return nil
// }

// func (c *Client) BroadcastGatewayCommand(command models.Command) error {
// 	topic := pubsub.GetMQTTGatewayResponseBroadcast("")

// 	err := pubsub.PublishMQTT(
// 		c.Ctx,
// 		c.MQTTClient,
// 		pubsub.MQTTPubOpts{
// 			Topic: topic,
// 			QoS:   1,
// 		},
// 		models.ContentJSON,
// 		command,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("failed to send command to topic %s: %w", topic, err)
// 	}
// 	return nil
// }

// // ========================
// // Requests
// // ========================

// func (c *Client) RequestGatewayIdentifiers(
// 	clientID string,
// 	filters map[string]interface{},
// 	reason string,
// ) error {
// 	cmd := models.Command{
// 		Name:   "request",
// 		Sender: c.Cfg.ID,
// 		Args: models.RequestArgs{
// 			Filters:   filters,
// 			Timestamp: time.Now(),
// 			Reason:    reason,
// 		},
// 	}
// 	if clientID == "+" || clientID == "" {
// 		return c.BroadcastGatewayCommand(cmd)
// 	}
// 	return c.SendGatewayCommand(clientID, cmd)
// }

// // ========================
// // Handlers
// // ========================

// func (c *Client) HandlerClientCommands(command models.Command) {
// 	switch command.Name {
// 	case "request":
// 		err := c.Client2GatewayResponse(
// 			command.Sender,
// 			models.Response{
// 				Name:      c.Cfg.ID,
// 				Status:    "",
// 				Data:      nil,
// 				Timestamp: time.Now(),
// 			},
// 		)
// 		if err != nil {
// 			log.Printf("Failed to send response: %v", err)
// 		}
// 		log.Printf("received request command: %v", command.Args)
// 	default:
// 		break
// 	}

// }
