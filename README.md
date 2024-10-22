# Video Processing Service

This service receives messages from a RabbitMQ queue indicating that new videos need processing. Upon receiving a message, it accesses a folder in the file system (simulating S3 storage) containing chunks of the video, merges these chunks to recreate the original video, converts it to MPEG Dash format, and updates the database to indicate successful conversion.

## Features

- **Message Handling:** Listens for messages on a RabbitMQ queue about new videos to process.
- **Chunk Merging:** Recreates the original video from chunks stored in a folder.
- **Video Conversion:** Converts the recreated video to MPEG Dash format.
- **Database Update:** Records the successful conversion in the database.

## Prerequisites

## Installation

## Usage

## Contributing
Feel free to submit issues, fork the repo, and submit pull requests.

## License
This project is licensed under the MIT License.