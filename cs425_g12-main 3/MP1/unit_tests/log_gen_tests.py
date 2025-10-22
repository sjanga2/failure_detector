# this unit test will call the generate_logs.py script on all servers to generate log files
# and then check if the log files are generated
import subprocess
import sys

# list of all vm ips and their hostnames
vm_list = ["172.22.94.224", "172.22.154.39", "172.22.158.39", "172.22.94.225", "172.22.154.40",
		"172.22.158.40", "172.22.94.226", "172.22.154.41", "172.22.158.41", "172.22.94.227"]

vm_to_hostname = {"172.22.94.224" : "fa25-cs425-1201.cs.illinois.edu",
                  "172.22.154.39" : "fa25-cs425-1202.cs.illinois.edu",
                  "172.22.158.39" : "fa25-cs425-1203.cs.illinois.edu",
                  "172.22.94.225" : "fa25-cs425-1204.cs.illinois.edu",
                  "172.22.154.40" : "fa25-cs425-1205.cs.illinois.edu",
                  "172.22.158.40" : "fa25-cs425-1206.cs.illinois.edu",
                  "172.22.94.226" : "fa25-cs425-1207.cs.illinois.edu",
                  "172.22.154.41" : "fa25-cs425-1208.cs.illinois.edu",
                  "172.22.158.41" : "fa25-cs425-1209.cs.illinois.edu",
                  "172.22.94.227" : "fa25-cs425-1210.cs.illinois.edu"}

# username = "sjanga2"
script_path = "~/cs425_mp1_g12/generate_logs.py"

def test_log_generation(username):
    # fixed parameters for the logs, can be changed
    n_lines = 1000
    rare_count = 2
    med_count = int(0.01 * n_lines)
    frequent_count = int(0.03 * n_lines)
    specific_count = 10
    odd_count = int(0.005 * n_lines)
    even_count = int(0.005 * n_lines)

    # loop through all VMs, and run the log gen script with the parameters
    for vm_ip in vm_list:
        hostname = vm_to_hostname[vm_ip]
        script_args = f"--rare_count {rare_count} --med_count {med_count} --frequent_count {frequent_count} --specific_count {specific_count} --odd_count {odd_count} --even_count {even_count} {n_lines} {hostname}"
        
        # ssh and run the script
        cmd = f"ssh {username}@{vm_ip} 'python3 {script_path} {script_args}'"
        print(cmd)
        subprocess.run(cmd, shell=True, capture_output=True, text=True)

        # check if the log file is created
        machine_id = hostname[13:15]
        remote_file = f"/home/{username}/cs425_mp1_g12/machine.{machine_id}.log"
        check_cmd = f"ssh {username}@{vm_ip} 'test -f {remote_file} && echo FOUND || echo NOT_FOUND'" # test -f on linux to check if file exists
        check_result = subprocess.run(check_cmd, shell=True, capture_output=True, text=True)

        if "FOUND" in check_result.stdout:
            print(f"Log file generated successfully on {hostname}")
        else:
            print(f"Log file generation failed on {hostname}")


if __name__ == "__main__":

    username = sys.argv[1]
    test_log_generation(username)