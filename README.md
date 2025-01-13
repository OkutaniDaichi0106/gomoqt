# gomoqt  
gomoqt is an original implementation of Media over QUIC in Golang, based on MOQTransport and MOQTransfork.  

## Features and Differences
### 1. Streams  
Streams completely based on MOQTranfork draft04.  

### 2. Messages  
Messages primarily based on MOQTransfork, some messages and fields are derived from MOQTransport  
#### Differences from MOQTransfork draft04  
- GOAWAY messages are transmitted on the Session Stream.  
- GROUP messages have a Publisher Priority field.  
- Two types of priorities: Subscriber Priority and Publisher Priority are implemented  
- No Track Path Parts field  




## Implemented sections
| Section                                      | Implemented        | Tested             |
| -------------------------------------------- | ------------------ | ------------------ |
| **2. Data Model**                            |                    |                    |
| 2.1. Frame                                   | :white_check_mark: | :x:                |
| 2.3. Group                                   | :white_check_mark: | :x:                |
| 2.4. Track                                   | :white_check_mark: | :x:                |
| 2.4.1. Track Naming and Scopes               | :construction:     | :x:                |
| 2.4.2. Scope                                 | :construction:     | :x:                |
| 2.4.3. Connection URL                        | :construction:     | :x:                |
| **3. Sessions**                              |                    |                    |
| 3.1. Session establishment                   | :white_check_mark: | :x:                |
| 3.1.1. WebTransport                          | :white_check_mark: | :x:                |
| 3.1.2. QUIC                                  | :white_check_mark: | :x:                |
| 3.2. Version and Extension Negotiation       | :white_check_mark: | :x:                |
| 3.3. Session initialization                  | :white_check_mark: | :x:                |
| 3.4. Stream Cancellation                     | :construction:     | :x:                |
| 3.5. Termination                             | :white_check_mark: | :x:                |
| 3.6. Migration                               | :construction:     | :x:                |
| **4. Data Transmittions**                    |                    |                    |
| 4.1 Publisher Priority Control               | :construction:     | :x:                |
| 4.2 Subscriber Priority Control              | :construction:     | :x:                |
| 4.3 Group Order Control                      | :construction:     | :x:                |
| 4.4 Cache                                    | :construction:     | :x:                |
| **5. Relays**                                |                    |                    |
| 5.1. Subscriber Interactions                 | :white_check_mark: | :x:                |
| 5.1.1. Graceful Publisher Relay Switchover   | :x:                | :x:                |
| 5.2. Publisher Interactions                  | :white_check_mark: | :x:                |
| 5.2.1. Graceful Publisher Network Switchover | :x:                | :x:                |
| 5.2.2. Graceful Publisher Relay Switchover   | :x:                | :x:                |
| 5.3. Relay Object Handling                   | :construction:     | :x:                |
| **Control Streams**                          |                    |                    |
| 6.1. Session Stream                          | :white_check_mark: | :white_check_mark: |
| 6.2. Announce Stream                         | :white_check_mark: | :white_check_mark: |
| 6.3. Subscribe Stream                        | :white_check_mark: | :white_check_mark: |
| 6.4. Fetch Stream                            | :white_check_mark: | :white_check_mark: |
| 6.5. Info Stream                             | :white_check_mark: | :white_check_mark: |
| **Control Messages**                         |                    |                    |
| 6.1. Parameters                              | :construction:     | :x:                |
| 6.1.1. Version Specific Parameters           | :white_check_mark: | :x:                |
| 6.2. SESSION_CLIENT                          | :white_check_mark: | :x:                |
| 6.2. SESSION_SERVER                          | :white_check_mark: | :x:                |
| 6.2. SESSION_UPDATE                          | :construction:     | :x:                |
| 6.2.1. Versions                              | :white_check_mark: | :x:                |
| 6.2.2. Setup Parameters                      | :white_check_mark: | :x:                |
| 6.3. GOAWAY                                  | :white_check_mark: | :x:                |
| 6.4. ANNOUNCE_INTEREST                       | :white_check_mark: | :x:                |
| 6.5. ANNOUNCE                                | :white_check_mark: | :x:                |
| 6.6. SUBSCRIBE                               | :white_check_mark: | :x:                |
| 6.7. SUBSCRIBE_UPDATE                        | :white_check_mark: | :x:                |
| 6.12. SUBSCRIBE_GAP                          | :construction:     | :x:                |
| 6.8. INFO                                    | :white_check_mark: | :x:                |
| 6.9. INFO_REQUEST                            | :white_check_mark: | :x:                |
| 6.10. FETCH                                  | :white_check_mark: | :x:                |
| 6.11. FETCH_UPDATE                           | :white_check_mark: | :x:                |
| **Data Stream**                              |                    |                    |
| 7.3. Group Streams                           | :white_check_mark: | :white_check_mark: |
| **Data Message**                             |                    |                    |
| 7.2. GROUP                                   | :white_check_mark: | :x:                |
| 7.2. FRAME                                   | :white_check_mark: | :x:                |
| **Datagram**                                 |                    |                    |
| 7.2. Datagram                                | :white_check_mark: | :x:                |
| **Security Considerations**                  |                    |                    |
| 8.1. Resource Exhaustion                     | :x:                | :x:                |

## Interoperablity test
We haven't conducted interoperability testing with other implementations yet

## TODO
- ANNOUNCE message's Status field
- Priority Control
- Handling the Group Expires
- LOC (Low Overhead Container)
- Common Catalog Format for moq