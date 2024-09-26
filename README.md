# gomoqu  
gomoqu is an implementation of Media over QUIC Transport in Golang.


| Section                                      | Implemented        | Tested     |
| -------------------------------------------- | ------------------ | ---------- |
| **Object Model**                             |                    |            |
| 2.1. Objects                                 | :white_check_mark: | :x:        |
| 2.2. Subgroups                               | :white_check_mark: | :x:        |
| 2.3. Groups                                  | :white_check_mark: | :x:        |
| 2.4. Track                                   | :white_check_mark: | :x:        |
| 2.4.1. Track Naming and Scopes               | :construction:     | :x:        |
| 2.4.2. Scope                                 | :construction:     | :x:        |
| 2.4.3. Connection URL                        | :construction:     | :x:        |
| **Sessions**                                 |                    |            |
| 3.1. Session establishment                   | :white_check_mark: | :x:        |
| 3.1.1. WebTransport                          | :white_check_mark: | :x:        |
| 3.1.2. QUIC                                  | :white_check_mark: | :x:        |
| 3.2. Version and Extension Negotiation       | :x:                | :x:        |
| 3.3. Session initialization                  | :white_check_mark: | :x:        |
| 3.4. Stream Cancellation                     | :x:                | :x:        |
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
| 6.4.1. Filter Types                          | :white_check_mark: | :x:        |
| 6.5. SUBSCRIBE_UPDATE                        | :white_check_mark: | :x:        |
| 6.6. UNSUBSCRIBE                             | :white_check_mark: | :x:        |
| 6.7. ANNOUNCE_OK                             | :white_check_mark: | :x:        |
| 6.8. ANNOUNCE_ERROR                          | :white_check_mark: | :x:        |
| 6.9. ANNOUNCE_CANCEL                         | :white_check_mark: | :x:        |
| 6.10. TRACK_STATUS_REQUEST                   | :construction:     | :x:        |
| 6.11. SUBSCRIBE_NAMESPACE                    | :white_check_mark: | :x:        |
| 6.12. UNSUBSCRIBE_NAMESPACE                  | :white_check_mark: | :x:        |
| 6.13. SUBSCRIBE_OK                           | :white_check_mark: | :x:        |
| 6.14. SUBSCRIBE_ERROR                        | :white_check_mark: | :x:        |
| 6.15. SUBSCRIBE_DONE                         | :white_check_mark: | :x:        |
| 6.16. MAX_SUBSCRIBE_ID                       | :x:                | :x:        |
| 6.17. ANNOUNCE                               | :white_check_mark: | :x:        |
| 6.18. UNANNOUNCE                             | :white_check_mark: | :x:        |
| 6.19. TRACK_STATUS                           | :construction:     | :x:        |
| 6.20. SUBSCRIBE_NAMESPACE_OK                 | :white_check_mark: | :x:        |
| 6.21. SUBSCRIBE_NAMESPACE_ERROR              | :white_check_mark: | :x:        |
| **Data Streams**                             |                    |            |
| 7.1. Object Headers                          | :white_check_mark: | :x:        |
| 7.2. Object Datagram Message                 | :white_check_mark: | :x:        |
| 7.3. Streams                                 | :white_check_mark: | :x:        |
| 7.3.1. Stream Header Track                   | :white_check_mark: | :x:        |
| 7.3.2. Stream Header Subgroup                | :white_check_mark: | :x:        |
| **Security Considerations**                  |                    |            |
| 8.1. Resource Exhaustion                     | :x:                | :x:        |
