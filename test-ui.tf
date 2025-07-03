resource "aws_instance" "test" {
  ami           = "ami-12345678"
  instance_type = "t2.micro"

  tags = {
    Name = "Test Instance"
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "test-bucket-12345"

  tags = {
    Environment = "dev"
  }
}