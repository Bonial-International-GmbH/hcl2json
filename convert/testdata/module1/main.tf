resource "aws_security_group" "sg" {
  name   = var.name
  vpc_id = var.vpc_id
}

resource "aws_security_group_rule" "egress-all" {
  cidr_blocks       = ["0.0.0.0/0"]
  description       = "Allow all egress traffic"
  from_port         = -1
  protocol          = "all"
  security_group_id = aws_security_group.sg.id
  to_port           = -1
  type              = "egress"
}
