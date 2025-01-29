echo "---- Client logs ----"
kubectl -n workload logs -l app=greeter-client
echo ""


echo "---- Server logs ----"
kubectl -n workload logs -l app=greeter-server
echo ""
