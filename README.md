# Tutorial SPIFFE SPIRE Kubernetes

> **Note:** Looking for the English version of this tutorial? [Click here](README_en.md)

Este tutorial irá guiá-lo através do processo de deploy do SPIRE em um cluster Kubernetes (kind) e usá-lo para emitir e gerenciar identidades SPIFFE para workloads em execução no cluster.

Este é um tutorial avançado para implantar o SPIRE em um ambiente Kubernetes, se você acabou de começar a aprender sobre
SPIRE, eu recomendo que você comece com o [guia de introdução ao SPIRE](https://spiffe.io/docs/latest/try/getting-started-linux-macos-x/) e
depois volte para este tutorial.

Não se esqueça de verificar estes outros repositórios com tutoriais e exemplos incríveis para explorar outros casos de uso do SPIFFE/SPIRE
como **integração com service meshes**, **bancos de dados**, **OIDC**, **federation** e mais!
- https://github.com/spiffe/spire-tutorials
- https://github.com/spiffe/spire-examples

## Requisitos
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [helm](https://helm.sh/docs/intro/install/)
- [docker] (https://docs.docker.com/get-docker/)

## Arquitetura

### Componentes SPIRE

Existem várias maneiras de implantar e executar o SPIRE em um cluster Kubernetes. Neste tutorial, fazemos uso dos seguintes projetos para
facilitar a implantação e o gerenciamento do SPIRE em um cluster Kubernetes.
- [spire-controller-manager](https://github.com/spiffe/spire-controller-manager): Usado para registro automático de identidades.
- [spire-helm-charts](https://github.com/spiffe/helm-charts-hardened): Usado para deploy de componentes SPIRE no cluster Kubernetes.
- [spiffe-csi](https://github.com/spiffe/spiffe-csi/tree/main): Usado para expor a Workload API em cada pod sem usar o hostPath volume.

Note que você não precisa desses projetos para executar o SPIRE no Kubernetes, mas eles são altamente recomendados, pois foram construídos pela comunidade para
uso em ambientes de produção.

Abaixo você pode encontrar um diagrama do que está sendo implantado no cluster sob o namespace `spire-system`:
- [spire-server](https://spiffe.io/docs/latest/spire-about/spire-concepts/#all-about-the-server): Implantado como um STS ao lado do contêiner spire-controller-manager.
- [spire-agent](https://spiffe.io/docs/latest/spire-about/spire-concepts/#all-about-the-agent): Implantado como um DaemonSet em cada nó do cluster.
- custom resources definitions: CRDs usados pelo spire-controller-manager para facilitar o registro de workloads.
- spire-spiffe-csi-driver: Implantado como um DaemonSet em cada nó do cluster para montar volumes da Workload API.

<img src="images/components.png" alt="drawing" width="1000"/>

### Componentes de Carga de Trabalho

Neste tutorial, usamos a aplicação `greeter` que pode ser encontrado no diretório `greeter`.
A aplicação greeter dispõe de um cliente e servidor que implementam o [exemplo de serviço hello world do gRPC](https://github.com/grpc/grpc-go/tree/master/examples).

O `hello world` é um serviço simples que retorna uma saudação ao cliente que o chama.

A ideia para o serviço greeter é demonstrar como reaver SVIDs da Workload API e usá-los para autenticar e autorizar a comunicação entre o cliente e o servidor.

Tanto o cliente quanto o servidor fazem uso da biblioteca [go-spiffe](https://github.com/spiffe/go-spiffe) para interagir com a Workload API e estabelecer uma conexão [mTLS](https://www.cloudflare.com/learning/access-management/what-is-mutual-tls/) entre eles.

<img src="images/workloads.png" alt="drawing" width="1000"/>

### Como as workloads obtêm seus SVIDs
Se você está usando o SPIRE sem o gerenciador de controle, você precisa passar pelo processo de [registro de identidade de workload](https://spiffe.io/docs/latest/deploying/registering/)
antes que suas workloads atestadas possam obter suas identidades.

Neste exemplo, estamos usando o spire-controller-manager com [ClusterSPIFFEID CR](https://github.com/spiffe/spire-controller-manager?tab=readme-ov-file#clusterspiffeid) para definir um
modelo para registrar automaticamente workloads **atestadas** com o SPIRE.

Este é o modelo que é usado em conjunto com [k8s workload attestor](https://github.com/spiffe/spire/blob/v1.8.7/doc/plugin_agent_workloadattestor_k8s.md):

`spiffe://{{ .TrustDomain }}/ns/{{ .PodMeta.Namespace }}/sa/{{ .PodSpec.ServiceAccountName }}`


### Passos de Implantação

0. Verificando requisitos.
   Este script verificará todas as ferramentas e dependências necessárias para executar o tutorial.
```bash
./0-check-requirements.sh
```

1. Criar um cluster kind.
   O segundo script criará um contêiner de registro para armazenar as imagens do serviço greeter.
   Ele também cria um cluster kind com dois nós.
```bash
./1-create-kind-cluster.sh
```

2. Deploy do SPIRE.
   Este script implantará os componentes spire no cluster kind usando o gráfico de helm spire.
```bash
./2-deploy-spire.sh
```

3. Deploy dos workloads de demonstração.
   Este script irá buildar e fazer o deploy do servidor greeter e o cliente no namespace "workload".
```bash
./3-deploy-demo-workloads.sh
```

4. Verificar que tudo está funcionando.
   Este script verificará que o cliente greeter pode se comunicar com o servidor greeter usando os SVIDs emitidos pelo SPIRE.
   Você poderá esperar alguns segundos para que os workloads sejam atestados e obtenham seus SVIDs.
```bash
./4-verify-demo-workloads.sh
```
 Se tudo estiver funcionando corretamente, você deverá ver a seguinte saída:
```bash
---- Client logs ----
2024/02/23 03:03:31 Starting up...
2024/02/23 03:03:31 Server Address: greeter-server.workload.svc.cluster.local:8443
2024/02/23 03:03:31 Connecting to Workload API at "unix:///spiffe-workload-api/spire-agent.sock"...
2024/02/23 03:03:37 Connected to Workload API at "unix:///spiffe-workload-api/spire-agent.sock"
2024/02/23 03:03:37 SPIFFE ID: "spiffe://example.org/ns/workload/sa/greeter-client-sa"
2024/02/23 03:03:37 Issuing requests every 30s...
2024/02/23 03:03:37 spiffe://example.org/ns/workload/sa/greeter-server-sa said "On behalf of spiffe://example.org/ns/workload/sa/greeter-client-sa, hello Joe!"

---- Server logs ----
2024/02/23 03:02:32 Starting up...
2024/02/23 03:02:32 Connecting to Workload API at "unix:///spiffe-workload-api/spire-agent.sock"...
2024/02/23 03:02:36 Connected to Workload API at "unix:///spiffe-workload-api/spire-agent.sock"
2024/02/23 03:02:36 SPIFFE ID: "spiffe://example.org/ns/workload/sa/greeter-server-sa"
2024/02/23 03:02:36 Serving on [::]:8443
2024/02/23 03:03:37 spiffe://example.org/ns/workload/sa/greeter-client-sa has requested that I say say hello to "Joe"...
```
Se você ver a saída acima, significa que tanto o cliente quanto o servidor foram atestados pelo SPIRE, receberam seus SVIDs e estão se comunicando entre si usando mTLS.

5. Limpar.
   Este script limpará o cluster kind e o contêiner de registro.
```bash
./5-clean-up.sh
```

## Referências e Recursos Adicionais
Caso você deseja saber mais sobre o SPIFFE e SPIRE,
