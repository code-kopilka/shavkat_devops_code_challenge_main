terraform {
	backend "local" {
		path = "terraform.tfstate"
	}
}

resource "aws_vpc" "main" {
	cidr_block = "10.1.0.0/16"

	tags = {
		Name        = "devops-code-challenge-vpc"
		Environment = "development"
		Project     = "devops-code-challenge"
	}
}

resource "aws_internet_gateway" "igw" { 
	vpc_id = aws_vpc.main.id

	tags = {
    Name = "main"
  }
}

variable "public_subnet_cidrs" {
	type        = list(string)
	description = "public subnets"
	default     = ["10.1.204.0/22", "10.1.208.0/22", "10.1.212.0/22"]
}

variable "private_subnet_cidrs" {
	type        = list(string)
	description = "private subnets"
	default     = ["10.1.104.0/22", "10.1.108.0/22", "10.1.112.0/22"]
}

resource "aws_subnet" "public_subnets" {
	count      = length(var.public_subnet_cidrs)
 	vpc_id     = aws_vpc.main.id
	cidr_block = element(var.public_subnet_cidrs, count.index)

	tags = {
		Name        = "public-${count.index + 1}"
		Environment = "development"
		Project     = "devops-code-challenge"
 	}
}

resource "aws_subnet" "private_subnets" {
	count      = length(var.private_subnet_cidrs)
	vpc_id     = aws_vpc.main.id
	cidr_block = element(var.private_subnet_cidrs, count.index)

	tags = {
		Name        = "private-${count.index + 1}"
		Environment = "development"
		Project     = "devops-code-challenge"
	}
}

resource "aws_route_table" "igw" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
	}

  tags = {
    Name        = "igw"
    Environment = "development"
    Project     = "devops-code-challenge"
  }
}

resource "aws_route_table_association" "public_subnet_igw" {
	count 					= length(var.public_subnet_cidrs)
	subnet_id				= element(aws_subnet.public_subnets[*].id, count.index)
	route_table_id	= aws_route_table.igw.id
}

resource "aws_security_group" "ec2" {
	name        = "devops-code-challenge-ec2"
	description = "Security group for devops code challenge application"
	vpc_id      = aws_vpc.main.id

	# Allow HTTP traffic from ALB or specific CIDR (restrict in production)
	ingress {
		description = "HTTP"
		from_port   = 3000
		to_port     = 3000
		protocol    = "tcp"
		cidr_blocks = ["10.1.0.0/16"] # Restrict to VPC only
	}

	# Allow SSH from specific IP (adjust for your use case)
	# ingress {
	#   description = "SSH"
	#   from_port   = 22
	#   to_port     = 22
	#   protocol    = "tcp"
	#   cidr_blocks = ["YOUR_IP/32"]
	# }

	egress {
		description = "All outbound traffic"
		from_port   = 0
		to_port     = 0
		protocol    = "-1"
		cidr_blocks = ["0.0.0.0/0"]
	}

	tags = {
		Name        = "devops-code-challenge-ec2"
		Environment = "development"
		Project     = "devops-code-challenge"
	}
}

resource "random_shuffle" "subnet" {
  input        = aws_subnet.public_subnets[*].id
  result_count = 1
}

# Use data source to find latest AMI dynamically
data "aws_ami" "amazon_linux" {
	most_recent = true
	owners      = ["amazon"]

	filter {
		name   = "name"
		values = ["al2023-ami-*-arm64"]
	}

	filter {
		name   = "virtualization-type"
		values = ["hvm"]
	}
}

resource "aws_instance" "box" {
	ami                         = data.aws_ami.amazon_linux.id
	instance_type               = "t4g.nano"
	subnet_id                   = random_shuffle.subnet.result[0]
	vpc_security_group_ids      = [aws_security_group.ec2.id]
	associate_public_ip_address = true

	# Add EBS volume for database persistence
	root_block_device {
		volume_type = "gp3"
		volume_size = 20
		encrypted   = true
	}

	tags = {
		Name        = "devops-code-challenge"
		Environment = "development"
		Project     = "devops-code-challenge"
	}

	depends_on = [
		aws_internet_gateway.igw
	]
}

resource "aws_s3_bucket" "example" {
	bucket = "devops-code-challenge-${random_id.bucket_suffix.hex}"

	tags = {
		Name        = "devops-code-challenge"
		Environment = "development"
		Project     = "devops-code-challenge"
	}
}

# Generate random suffix to ensure bucket name uniqueness
resource "random_id" "bucket_suffix" {
	byte_length = 4
}

# Enable versioning
resource "aws_s3_bucket_versioning" "example" {
	bucket = aws_s3_bucket.example.id
	versioning_configuration {
		status = "Enabled"
	}
}

# Enable encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "example" {
	bucket = aws_s3_bucket.example.id

	rule {
		apply_server_side_encryption_by_default {
			sse_algorithm = "AES256"
		}
	}
}

# Block public access
resource "aws_s3_bucket_public_access_block" "example" {
	bucket = aws_s3_bucket.example.id

	block_public_acls       = true
	block_public_policy     = true
	ignore_public_acls      = true
	restrict_public_buckets = true
}
