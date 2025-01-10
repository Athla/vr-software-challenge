Your task is to build an HTTP REST application in Java or Golang that meets the requirements described below. Beyond the specified requirements, as well as any language limitations indicated by the technical implementation notes below and/or hiring managers, the application is of your own technical authorship.
Requirement 1: Store a Purchase Transaction
Your application must be capable of accepting and storing (i.e., persisting) a purchase transaction with the following attributes:

    Description: A brief description of the purchase.
    Transaction Date: The date of the transaction.
    Purchase Amount: The amount in U.S. dollars.

When stored, the transaction must be assigned a unique identifier.
Field Requirements

    Description: Must not exceed 50 characters.
    Transaction Date: Must be a valid date format.
    Purchase Amount: Must be a valid positive value, rounded to the nearest cent.
    Unique Identifier: Must uniquely identify the purchase.

Optional Requirements (Considered a Plus)
The application should process this transaction through a queue (e.g., RabbitMQ, Kafka, or similar). This means that input validations will be processed synchronously, but the transaction should be persisted asynchronously.
