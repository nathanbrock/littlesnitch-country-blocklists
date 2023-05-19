import json

def flush_to_file(rules, output_file):
    output = {
        "description": "",
        "name": "Test",
        "rules": rules
    }

    with open(output_file, 'w') as file:
        file.write(json.dumps(output, indent=2))


def cidr_to_littlesnitch_file(input_file, output_file):
    with open(input_file, 'r') as file:
        cidr_list = file.read().splitlines()

    rules = []
    flush = 1
    for cidr in cidr_list:
        ip, mask = cidr.split('/')
        rule = {
            "action": "deny",
            "priority": "high",
            "remote-addresses": f"{ip}/{mask}"
        }
        rules.append(rule)

        if len(rules) >= 200000:
            flush_to_file(rules, str(flush) + "_" + output_file)
            flush = flush + 1
            rules = []

    flush_to_file(rules, str(flush) + "_" + output_file)


# Example usage
input_file = "firewall.txt"
output_file = "littlesnitch_rules.lsrules"
cidr_to_littlesnitch_file(input_file, output_file)
