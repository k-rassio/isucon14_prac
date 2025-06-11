#!/bin/bash
set -eux

isu_repo_dir="/home/user/isucon14"
isu_webapp_dir="./webapp/"
isu_ansible_dir="/home/user/isucon14/provisioning/ansible"
webapp_tar_dir="/home/user/isucon14/provisioning/ansible/roles/webapp/files/webapp.tar.gz"

cd $isu_repo_dir

echo "############# git commit & push ###############"
# git add . || true
# git commit -m "$1_`date "+%Y_%m_%d_%H_%M_%S"`" || true
git push -u origin || true

echo "############ appサーバへのデプロイ ################"
tar -zcvf $webapp_tar_dir $isu_webapp_dir
cd $isu_ansible_dir
ansible-playbook -i inventory/inventory.yml -l isu14_app application-deploy.yml

echo "################ benchの実行 ####################"
ansible -i inventory/inventory.yml isu14_bench -b --become-user isucon -a "/home/isucon/bench_linux_amd64 run ./bench_linux_amd64 run --target=https://192.168.37.134 --payment-url=http://192.168.37.134:12345" | tee ../../result/bench_result/bench_`date +%m%d%H%M`.log 

echo "################ benchの結果を取得 ####################"

echo "############### mysqldumpの実行 #################"
ansible-playbook -i inventory/inventory.yml -l isu14_app mysql_slowquery.yml --skip-tags=install
echo "################# alpの実行 #####################"
ansible-playbook -i inventory/inventory.yml -l isu14_app alp.yml
