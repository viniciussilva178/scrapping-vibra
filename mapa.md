# Queries

[x] - 
## Pegar Linha digitavel em `public_debito_direto_autorizado`
```sql
    select * 
    from public.debito_direto_autorizado 
    where linha_digitavel= $1;
```

[x] - 
 ## Popular `sis.documento`
 ### Pegar empresa
```sql
    select	se.id_empresa
    from	pessoa p 
        inner join vibra.conta_pagar cp on (cp.cnpj_pagador = p.cnpj_cpf)
        inner join public.sis_empresa se on (se.id_pessoa = p.id_pessoa)
```

[x] -
### Pegar nome do documento (nome_documento = numero_documento)
```sql
    select numero_documento
    from  vibra.conta_pagar p
    where linha_digitavel='03399514854562510126157281001016111780005494600'
```

[x] -
### Pegar documento (conteudo em bytes)
```sql
    select conteudo_documento
    from vibra.documento_abastecimento p 
        inner join vibra.conta_pagar cp on (cp.id_nota_fiscal = p.id_documento_abastecimento)
        inner join public.debito_direto_autorizado d on (d.linha_digitavel = cp.linha_digitavel)
```

[x] -
### Pegar Tamanho do Documento (length)
```sql 
    select bit_length(conteudo_documento)
    from vibra.documento_abastecimento p 
        inner join vibra.conta_pagar cp on (cp.id_nota_fiscal = p.id_documento_abastecimento)
        inner join public.debito_direto_autorizado d on (d.linha_digitavel = cp.linha_digitavel)
```
insert into public.sis_relacao_registro 
(id_chave_registro_a, id_chave_registro_b, tipo_relacao, id_usuario_inclusao) 
values (3186167, 27898, 'sis_documento_debito_direto_autorizado', 53);

[x] -
### Pegar tipo do Documento

[x] -
### Insertar dentro de `public.sis_documento` 
```sql
        insert into public.sis_documento
    (id_empresa, nome_documento, tamanho, id_usuario_inclusao, id_tipo_documento) 
    values ($1, $2, $3, 53, $4);
```

[x]- 
### Insertar em `public.sis_relacao_registro`
    1 - id_chave_registro_a = id_documento em (public.sis_documento)
    2 - id_chave_registro__b =  id_debito_direto_autorizado em (public_debito_direto_autorizado)
    
```sql
    insert into public.sis_relacao_registro 
    (id_chave_registro_a, id_chave_registro_b, tipo_relacao, id_usuario_inclusao) 
    values ($1, $2, 'sis_documento_debito_direto_autorizado', 53);
```
