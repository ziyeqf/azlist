# azlist

List Azure resources by an Azure Resource Graph [`where` predicate](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/whereoperator).

## Example

```
azlist 'resourceGroup =~ "example-rg"'
```

## FAQ

- **Question**: What is the difference of the resource list returned by `azlist` and ARG?
    
    **Answer**: By default, they are the same. While if `azlist` is called with `--recursive`, it returns more than ARG. The ARG only returns ARM tracked resources, but not for the RP proxy resources (e.g. subnet, network security rules, storage containers, etc). `azlist --recursive` will return all the tracked and proxy resources.

- **Question**: What is the difference of the resource list returned by `azlist --recursive` and ARM template export?
    
    **Answer**: They are meant to be the same. But ARM template export only support some certain falvors (e.g. resource group), while `azlist` allows more. However, `azlist` returns less information for each resource, e.g. it doesn't have the cross resource dependency.

- **Question**: Why predicate `type =~ "microsoft.network/virtualnetworks/subnets"` returns me nothing, even with `--recursive`?
    
    **Answer**: This is because `azlist` will first make an ARG call with the given `where` predicate, then if `--recursive` is specified, it will recursively call the "LIST" on the *known* child resource types. In this case, since the subnet is not an ARM tracked resource, ARG returns nothing.



