
# import mp1_utils
import subprocess
import os


def run_go_client(command):
    # os.path doesn't recognize ~ so need to explicitly expand user to /home/username
    client_dir = os.path.expanduser("~/cs425_mp1_g12/client")

    # run the clinent process
    process = subprocess.Popen(
        ["go", "run", "client.go", "unit_test"],
        cwd=client_dir,               
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True, # stdin etc are strings 
    )

    #sends the grep command to client
    process.stdin.write(command + "\n")
    #sends the exit command client
    process.stdin.write("exit\n")   
    process.stdin.flush()

    # collect output and waits for client to exit
    stdout, stderr = process.communicate()
    return stdout, stderr

def test_cmd_rare_all(out, err):
    #check that err is empty
    assert err == "", f"Error occurred: {err}"

    # split out by replies in each machine
    machine_blocks = out.split("Reply from ")

    # check each machine for rare pattern count
    for i, block in enumerate(machine_blocks):
        if i == 0:
            continue  
        # print(i, block)
        count = block.count("RARE_PATTERN_ALL_MACHINES")
        assert count == 2, f"Machine {i} has {count} occurrences, expected 2"
    print("test_cmd_rare_all: PASSED")


def test_cmd_med_all(out, err):
    #check that err is empty
    assert err == "", f"Error occurred: {err}"

    # split out by replies in each machine
    machine_blocks = out.split("Reply from ")

    # check each machine for rare pattern count
    for i, block in enumerate(machine_blocks):
        if i == 0:
            continue  
        # print(i, block)
        if i == 1:
            expected_count = 3000
        else:
            expected_count = 10
        count = block.count("MED_PATTERN_ALL_MACHINES")
        assert count == expected_count, f"Machine {i} has {count} occurrences, expected {expected_count}"
    print("test_cmd_med_all: PASSED")

def test_cmd_frequent_all(out, err):
    #check that err is empty
    assert err == "", f"Error occurred: {err}"

    # split out by replies in each machine
    machine_blocks = out.split("Reply from ")

    # check each machine for rare pattern count
    for i, block in enumerate(machine_blocks):
        if i == 0:
            continue  
        # print(i, block)
        if i == 1:
            expected_count = 9000
        else:
            expected_count = 30
        count = block.count("FREQUENT_PATTERN_ALL_MACHINES")
        assert count == expected_count, f"Machine {i} has {count} occurrences, expected {expected_count}"
    print("test_cmd_frequent_all: PASSED")

def test_cmd_only_on_machine_02(out, err):
    #check that err is empty
    assert err == "", f"Error occurred: {err}"

    # split out by replies in each machine
    machine_blocks = out.split("Reply from ")

    # check each machine for pattern count
    for i, block in enumerate(machine_blocks):
        if i == 0:
            continue
        count = block.count("PATTERN_ONLY_ON_MACHINE_2")
        if i == 2:
            assert count == 10, f"Machine {i} has {count} occurrences, expected 10"
        else:
            assert count == 0, f"Machine {i} has {count} occurrences, expected 0"
    print("test_cmd_only_on_machine_02: PASSED")

def test_cmd_only_on_machine_03(out, err):
    assert err == "", f"Error occurred: {err}"
    # split out by replies in each machine
    machine_blocks = out.split("Reply from ")

    # check each machine for pattern count
    for i, block in enumerate(machine_blocks):
        if i == 0:
            continue
        count = block.count("PATTERN_ONLY_ON_MACHINE_3")
        if i == 3:
            assert count == 10, f"Machine {i} has {count} occurrences, expected 10"
        else:
            assert count == 0, f"Machine {i} has {count} occurrences, expected 0"
    print("test_cmd_only_on_machine_03: PASSED")

def test_cmd_only_on_machine_07(out, err):
    assert err == "", f"Error occurred: {err}"
    # split out by replies in each machine
    machine_blocks = out.split("Reply from ")

    # check each machine for pattern count
    for i, block in enumerate(machine_blocks):
        if i == 0:
            continue
        count = block.count("PATTERN_ONLY_ON_MACHINE_7")
        if i == 7:
            assert count == 10, f"Machine {i} has {count} occurrences, expected 10"
        else:
            assert count == 0, f"Machine {i} has {count} occurrences, expected 0"
    print("test_cmd_only_on_machine_07: PASSED")

def test_cmd_only_on_odd_machine_nos(out, err):
    assert err == "", f"Error occurred: {err}"
    # split out by replies in each machine
    machine_blocks = out.split("Reply from ")

    # check each machine for pattern count
    for i, block in enumerate(machine_blocks):
        if i == 0:
            continue
        count = block.count("PATTERN_ONLY_ON_ODD_MACHINE_NOS")
        if i == 1:
            assert count == 1500, f"Machine {i} has {count} occurrences, expected 1500"
        elif i % 2 == 1:  # odd machine numbers
            assert count == 5, f"Machine {i} has {count} occurrences, expected 5"
        else:
            assert count == 0, f"Machine {i} has {count} occurrences, expected 0"
    print("test_cmd_only_on_odd_machine_nos: PASSED")

def test_cmd_only_on_even_machine_nos(out, err):
    assert err == "", f"Error occurred: {err}"
    # split out by replies in each machine
    machine_blocks = out.split("Reply from ")

    # check each machine for pattern count
    for i, block in enumerate(machine_blocks):
        if i == 0:
            continue
        count = block.count("PATTERN_ONLY_ON_EVEN_MACHINE_NOS")
        if i % 2 == 0:  # even machine numbers
            assert count == 5, f"Machine {i} has {count} occurrences, expected 5"
        else:
            assert count == 0, f"Machine {i} has {count} occurrences, expected 0"
    print("test_cmd_only_on_even_machine_nos: PASSED")

def test_cmd_nonexistent_pattern(out, err):
    assert err == "", f"Error occurred: {err}"
    # split out by replies in each machine
    machine_blocks = out.split("Reply from ")

    # check each machine for pattern count
    for i, block in enumerate(machine_blocks):
        if i == 0:
            continue
        count = block.count("NONEXISTENT_PATTERN_I_WILL_BE_SURPRISED_IF_YOU_FIND_ME")
        assert count == 0, f"Machine {i} has {count} occurrences, expected 0"
    print("test_cmd_nonexistent_pattern: PASSED")

def test_cmd_count_critical(out, err):
    assert err == "", f"Error occurred: {err}"
    # split out by replies in each machine
    machine_blocks = out.split("Reply from ")

    for i, block in enumerate(machine_blocks):
        if i == 0:
            continue
        assert ((block.strip()[-1]) == '2'), f"Machine {i} has incorrect count, expected 2"
    print("test_cmd_count_critical: PASSED")

def test_regex_odd_even(out,err):
    #check that err is empty
    assert err == "", f"Error occurred: {err}"

    # split out by replies in each machine
    machine_blocks = out.split("Reply from ")

    # check each machine for rare pattern count
    for i, block in enumerate(machine_blocks):
        if i == 0:
            continue  
        # print(i, block)
        if i == 1:
            expected_count = 1500
        else:
            expected_count = 5
        # can be either odd or even
        count = block.count("PATTERN_ONLY_ON_ODD_MACHINE_NOS")
        #adding even to count 
        count += block.count("PATTERN_ONLY_ON_EVEN_MACHINE_NOS")
        assert count == expected_count, f"Machine {i} has {count} occurrences, expected {expected_count}"
    print("test_regex_odd_even: PASSED")

def test_regex_2(out,err):
    #check that err is empty
    assert err == "", f"Error occurred: {err}"

    # split out by replies in each machine
    machine_blocks = out.split("Reply from ")

    # check each machine for rare pattern count
    
    for i, block in enumerate(machine_blocks):
        count = 0
        if i == 0:
            continue  
        # print(i, block)
        if i == 1:
            count += block.count("PATTERN_ONLY_ON_MACHINE_1")
        if i ==2:
            count += block.count("PATTERN_ONLY_ON_MACHINE_2")
        if i ==3:
            count += block.count("PATTERN_ONLY_ON_MACHINE_3")
        if i ==4:
            count += block.count("PATTERN_ONLY_ON_MACHINE_4")
        if i ==5:      
            count += block.count("PATTERN_ONLY_ON_MACHINE_5")
        if i ==6:
            count += block.count("PATTERN_ONLY_ON_MACHINE_6")
        if i ==7:
            count += block.count("PATTERN_ONLY_ON_MACHINE_7")
        if i ==8:
            count += block.count("PATTERN_ONLY_ON_MACHINE_8")
        if i ==9:
            count += block.count("PATTERN_ONLY_ON_MACHINE_9")
        if i ==10:
            count += block.count("PATTERN_ONLY_ON_MACHINE_10")
        # # can be either odd or even
        # count = block.count("PATTERN_ONLY_ON_MACHINE_01")
        # #adding even to count 
        # count += block.count("PATTERN_ONLY_ON_EVEN_MACHINE_NOS")
        assert count == 10, f"Machine {i} has {count} occurrences, expected {10}"
    print("test_regex_odd_even: PASSED")


def test_regex_3(out,err):
    #check that err is empty
    assert err == "", f"Error occurred: {err}"

    # split out by replies in each machine
    machine_blocks = out.split("Reply from ")

    # check each machine for rare pattern count
    
    for i, block in enumerate(machine_blocks):
        count = 0
        if i == 0:
            continue  
        # print(i, block)
        if i == 1:
            count += block.count("PATTERN_ONLY_ON_MACHINE_1")
        if i ==2:
            count += block.count("PATTERN_ONLY_ON_MACHINE_2")
        if i ==3:
            count += block.count("PATTERN_ONLY_ON_MACHINE_3")
        if i ==4:
            count += block.count("PATTERN_ONLY_ON_MACHINE_4")
        if i ==5:      
            count += block.count("PATTERN_ONLY_ON_MACHINE_5")
        if i ==6:
            count += block.count("PATTERN_ONLY_ON_MACHINE_6")
        if i ==7:
            count += block.count("PATTERN_ONLY_ON_MACHINE_7")
        if i ==8:
            count += block.count("PATTERN_ONLY_ON_MACHINE_8")
        if i ==9:
            count += block.count("PATTERN_ONLY_ON_MACHINE_9")
        if i ==10:
            count += block.count("PATTERN_ONLY_ON_MACHINE_10")

        if i  <5 or i == 10:
            assert count == 10, f"Machine {i} has {count} occurrences, expected {10}"
        else:
            assert count == 0, f"Machine {i} has {count} occurrences, expected {0}"
    print("test_regex_odd_even: PASSED")


if __name__ == "__main__":
    # grep commands to test on
    commands = [
        # known pattern test
        "grep RARE_PATTERN_ALL_MACHINES", # 0
        "grep MED_PATTERN_ALL_MACHINES", # 1
        "grep FREQUENT_PATTERN_ALL_MACHINES", # 2
        "grep PATTERN_ONLY_ON_MACHINE_2", # 3
        "grep PATTERN_ONLY_ON_MACHINE_3", # 4
        "grep PATTERN_ONLY_ON_MACHINE_7", # 5
        "grep PATTERN_ONLY_ON_ODD_MACHINE_NOS", # 6
        "grep PATTERN_ONLY_ON_EVEN_MACHINE_NOS", # 7
        "grep NONEXISTENT_PATTERN_I_WILL_BE_SURPRISED_IF_YOU_FIND_ME", # 8

        # regex test
        "grep -E \"PATTERN_ONLY_ON_(ODD|EVEN)?_MACHINE_NOS\"", # 9
        "grep -E \"PATTERN_ONLY_ON_MACHINE_.*\"", # 10
        "grep -E \"PATTERN_ONLY_ON_MACHINE_[0-4]+\"", # 11

        # other grep options
        "grep -i rare_pattern_all_machines", # 12
        "grep -c CRITICAL" # 13
        # "grep -c \"ODD_MACHINE\"",
        # "grep -E \"PATTERN_ONLY_ON_(ODD|EVEN)?_MACHINE_NOS\"",
        # "grep RARE_PATTERN"
    ]


    #test 1
    # print(f"Running command: {commands[0]}")
    out, err = run_go_client(commands[0])
    # print(out,err)
    test_cmd_rare_all(out, err)

    #test 2
    # print(f"Running command: {commands[1]}")
    out, err = run_go_client(commands[1])
    # print(out,err)
    test_cmd_med_all(out, err)
    #test 3
    # print(f"Running command: {commands[2]}")
    out, err = run_go_client(commands[2])
    # print(out,err)
    test_cmd_frequent_all(out, err)

    # print(f"Running command: {commands[3]}")
    out, err = run_go_client(commands[3])
    #print(out, err)
    test_cmd_only_on_machine_02(out, err)

    # print(f"Running command: {commands[4]}")
    out, err = run_go_client(commands[4])
    test_cmd_only_on_machine_03(out, err)

    # print(f"Running command: {commands[5]}")
    out, err = run_go_client(commands[5])
    test_cmd_only_on_machine_07(out, err)

    # print(f"Running command: {commands[6]}")
    out, err = run_go_client(commands[6])
    test_cmd_only_on_odd_machine_nos(out, err)

    # print(f"Running command: {commands[7]}")
    out, err = run_go_client(commands[7])
    test_cmd_only_on_even_machine_nos(out, err)

    # print(f"Running command: {commands[8]}")
    out, err = run_go_client(commands[8])
    test_cmd_nonexistent_pattern(out, err)

    #test 9
    # print(f"Running command: {commands[9]}")
    out, err = run_go_client(commands[9])
    # print(out,err)
    test_regex_odd_even(out, err)

    #test10
    # print(f"Running command: {commands[10]}")
    out, err = run_go_client(commands[10])
    # print(out,err)
    test_regex_2(out, err)

    #test11
    # print(f"Running command: {commands[11]}")
    out, err = run_go_client(commands[11])
    # print(out,err)
    test_regex_3(out, err)

    # print(f"Running command: {commands[12]}")
    out, err = run_go_client(commands[12])
    test_cmd_rare_all(out, err)

    # print(f"Running command: {commands[13]}")
    out, err = run_go_client(commands[13])
    test_cmd_count_critical(out, err)

    # print(f"Running command: {commands[9]}")
    # out, err = run_go_client(commands[9])
    # print(out,err)
    # print(f"Running command: {commands[10]}")
    # out, err = run_go_client(commands[10])
    # print(out,err)
    # print(f"Running command: {commands[11]}")
    # out, err = run_go_client(commands[11])
    # print(out,err)
    # test_cmd_rare_all(out, err)