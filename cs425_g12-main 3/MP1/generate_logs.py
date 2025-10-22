import logging
import random
import string
import os
import argparse

def generate_random_line(length=80):
    return ''.join(random.choices(string.ascii_letters + string.digits, k=length))

def insert_pattern(line, pattern):
    pos = random.randint(0, len(line) - len(pattern))
    return line[:pos] + pattern + line[pos + len(pattern):]

def generate_log(hostname: str, n_lines: int,
                 rare_count: int = 1, med_count: int = 10, frequent_count: int = 100,
                 odd_count: int = 5, even_count: int = 5,
                 specific_count: int = 5):

    machine_id = int(hostname.split('-')[2].split('.')[0][-2:])
    log_dir = os.path.expanduser("~/cs425_mp1_g12")
    filename = os.path.join(log_dir, f"machine.{machine_id:02d}.log")

    # python logging config
    logging.basicConfig(
        filename=filename,
        filemode='w',
        format='%(asctime)s %(levelname)s %(message)s',
        datefmt='%Y-%m-%d %H:%M:%S',
        level=logging.INFO
    )

    #patterns that will be included into log lines
    rare = "RARE_PATTERN_ALL_MACHINES"
    med = "MED_PATTERN_ALL_MACHINES"
    frequent = "FREQUENT_PATTERN_ALL_MACHINES"
    specific = f"PATTERN_ONLY_ON_MACHINE_{machine_id}"
    odd = "PATTERN_ONLY_ON_ODD_MACHINE_NOS"
    even = "PATTERN_ONLY_ON_EVEN_MACHINE_NOS"

    #randomize line nos for the patterns to occur on
    line_indices = list(range(n_lines))
    random.shuffle(line_indices)
    idx = 0

    rare_lines = line_indices[idx:idx + rare_count]
    idx += rare_count
    med_lines = line_indices[idx:idx + med_count]
    idx += med_count
    freq_lines = line_indices[idx:idx + frequent_count]
    idx += frequent_count
    specific_lines = line_indices[idx:idx + specific_count]
    idx += specific_count

    if machine_id % 2 == 1:
        odd_lines = line_indices[idx:idx + odd_count]
        idx += odd_count
        even_lines = []
    else:
        even_lines = line_indices[idx:idx + even_count]
        idx += even_count
        odd_lines = []

    for i in range(n_lines):
        line = generate_random_line()
        if i in freq_lines:
            # infor level
            line = insert_pattern(line, frequent)
            level = logging.INFO
        elif i in med_lines:
            #warning level
            line = insert_pattern(line, med)
            level = logging.WARNING
        elif i in rare_lines:
            # critical level
            line = insert_pattern(line, rare)
            level = logging.CRITICAL
        elif i in specific_lines or i in odd_lines or i in even_lines:
            # error levels
            if i in specific_lines:
                line = insert_pattern(line, specific)
            if i in odd_lines:
                line = insert_pattern(line, odd)
            if i in even_lines:
                line = insert_pattern(line, even)
            level = logging.ERROR
        else:
            # random line just with level as info
            level = logging.INFO
        logging.log(level, line)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Generate log files with patterns.")

    parser.add_argument("n_lines", type=int, help="Number of lines to generate")
    parser.add_argument("hostname", type=str, help="Hostname (e.g., fa23-cs425-01.cs.illinois.edu)")

    parser.add_argument("--rare_count", type=int, default=1, help="Number of rare pattern lines")
    parser.add_argument("--med_count", type=int, default=10, help="Number of medium pattern lines")
    parser.add_argument("--frequent_count", type=int, default=100, help="Number of frequent pattern lines")
    parser.add_argument("--odd_count", type=int, default=5, help="Number of odd machine pattern lines")
    parser.add_argument("--even_count", type=int, default=5, help="Number of even machine pattern lines")
    parser.add_argument("--specific_count", type=int, default=5, help="Number of machine-specific pattern lines")

    args = parser.parse_args()

    generate_log(
        hostname=args.hostname,
        n_lines=args.n_lines,
        rare_count=args.rare_count,
        med_count=args.med_count,
        frequent_count=args.frequent_count,
        odd_count=args.odd_count,
        even_count=args.even_count,
        specific_count=args.specific_count,
    )