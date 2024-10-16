# gomoqt  
gomoqt is an original implementation of Media over QUIC in Golang, refering to MOQTransport and MOQTransfork.  
[document](https://www.notion.so/gomoqt-116e4265c81c80f190aacad73bfdae5a?pvs=4)

## Implemented sections
| Section                                      | Implemented        | Tested     |
| -------------------------------------------- | ------------------ | ---------- |
| **Data Model**                             |                    |            |
| 2.1. Frame                                   | :white_check_mark: | :x:        |
| 2.3. Group                                   | :white_check_mark: | :x:        |
| 2.4. Track                                   | :white_check_mark: | :x:        |
| 2.4.1. Track Naming and Scopes               | :construction:     | :x:        |
| 2.4.2. Scope                                 | :construction:     | :x:        |
| 2.4.3. Connection URL                        | :construction:     | :x:        |
| **Sessions**                                 |                    |            |
| 3.1. Session establishment                   | :white_check_mark: | :x:        |
| 3.1.1. WebTransport                          | :white_check_mark: | :x:        |
| 3.1.2. QUIC                                  | :white_check_mark: | :x:        |
| 3.2. Version and Extension Negotiation       | :white_check_mark: | :x:        |
| 3.3. Session initialization                  | :white_check_mark: | :x:        |
| 3.4. Stream Cancellation                     | :construction:     | :x:        |
| 3.5. Termination                             | :white_check_mark: | :x:        |
| 3.6. Migration                               | :construction:     | :x:        |
| **Priorities**                               |                    |            |
| 4. Priorities                                | :construction:     | :x:        |
| **Relays**                                   |                    |            |
| 5.1. Subscriber Interactions                 | :white_check_mark: | :x:        |
| 5.1.1. Graceful Publisher Relay Switchover   | :x:                | :x:        |
| 5.2. Publisher Interactions                  | :white_check_mark: | :x:        |
| 5.2.1. Graceful Publisher Network Switchover | :x:                | :x:        |
| 5.2.2. Graceful Publisher Relay Switchover   | :x:                | :x:        |
| 5.3. Relay Object Handling                   | :construction:     | :x:        |
| **Control Messages**                         |                    |            |
| 6.1. Parameters                              | :construction:     | :x:        |
| 6.1.1. Version Specific Parameters           | :white_check_mark: | :x:        |
| 6.2. CLIENT_SETUP                            | :white_check_mark: | :x:        |
| 6.2. SERVER_SETUP                            | :white_check_mark: | :x:        |
| 6.2.1. Versions                              | :white_check_mark: | :x:        |
| 6.2.2. Setup Parameters                      | :white_check_mark: | :x:        |
| 6.3. GOAWAY                                  | :white_check_mark: | :x:        |
| 6.4. SUBSCRIBE                               | :white_check_mark: | :x:        |
| 6.5. SUBSCRIBE_UPDATE                        | :white_check_mark: | :x:        |
| 6.10. TRACK_STATUS_REQUEST                   | :construction:     | :x:        |
| 6.11. SUBSCRIBE_NAMESPACE                    | :white_check_mark: | :x:        |
| 6.17. ANNOUNCE                               | :white_check_mark: | :x:        |
| 6.19. TRACK_STATUS                           | :construction:     | :x:        |
| **Data Message**                             |                    |            |
| 7.2. GROUP                                   | :white_check_mark: | :x:        |
| 7.2. FRAME                                   | :white_check_mark: | :x:        |
| **Datagram**                                 |                    |            |
| 7.2. Datagram                                | :white_check_mark: | :x:        |
| **Data Stream**                              |                    |            |
| 7.3. Streams                                 | :white_check_mark: | :x:        |
| **Security Considerations**                  |                    |            |
| 8.1. Resource Exhaustion                     | :x:                | :x:        |

## Interoperablity test
We haven't conducted interoperability testing with other implementations yet