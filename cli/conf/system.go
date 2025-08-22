package conf


// switch msg.Action.Type {
// 	case CREATE_SERVER:
// 		{
// 			var globalConf core.Config
// 			err := msg.DecodePayload(&globalConf)
// 			if err != nil {
// 				errMsg := NewErrorMessage("Invalid config struct provided", err.Error())
// 				if msg.RequestID != "" {
// 					errMsg.RequestID = msg.RequestID
// 				}
// 				client.sendMsg(errMsg)
// 				return nil
// 			}

// 			for _, singleServer := range globalConf.Servers {
// 				s.globalConf.Servers = append(s.globalConf.Servers, singleServer)
// 				s.HostConfig[singleServer.Name] = createSingleHttpServer(singleServer)
// 			}

// 			if globalConf.HealthCheckInterval != 0 {
// 				s.globalConf.HealthCheckInterval = globalConf.HealthCheckInterval
// 			}
// 			if globalConf.LogOutput != "" {
// 				s.globalConf.LogOutput = globalConf.LogOutput
// 			}
// 			s.globalConf.Middlewares = globalConf.Middlewares
// 			if msg.RequestID != "" {
// 				newMsg, err := NewMessage(CREATE_SERVER, struct {
// 					config  *core.Config
// 					success bool
// 				}{config: s.globalConf, success: true}, nil)

// 				if err != nil {
// 					client.send <- &Message{RequestID: msg.RequestID}
// 					return nil
// 				}

// 				newMsg.RequestID = msg.RequestID

// 				client.sendMsg(newMsg)
// 			}
// 			return nil
// 		}
// 	case ROLLBACK_SERVER:
// 		{
// 			var server core.Server
// 			err := msg.DecodePayload(&server)
// 			if err != nil {
// 				errMsg := NewErrorMessage("No server provided for roll backing", err.Error())
// 				if msg.RequestID != "" {
// 					errMsg.RequestID = msg.RequestID
// 				}
// 				client.sendMsg(errMsg)
// 				return nil
// 			}

// 			for idx, confServer := range s.globalConf.Servers {
// 				if confServer.Name == server.Name {
// 					if server.BalancingServers != nil {
// 						s.globalConf.Servers[idx].RollBack(server.BalancingServers)
// 					}
// 					s.globalConf.Servers[idx] = &server
// 					s.HostConfig[confServer.Name] = createSingleHttpServer(&server)
// 				}
// 			}

// 			if msg.RequestID != "" {
// 				newMsg, err := NewMessage(CREATE_SERVER, struct {
// 					config  *core.Config
// 					success bool
// 				}{config: s.globalConf, success: true}, nil)

// 				if err != nil {
// 					client.send <- &Message{RequestID: msg.RequestID}
// 					return nil
// 				}

// 				newMsg.RequestID = msg.RequestID

// 				client.sendMsg(newMsg)
// 			}

// 			return nil
// 		}
// 	}
// 	return nil