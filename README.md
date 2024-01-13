# Golang REST API Specification 

This project shows how to generate an OpenAPI 2.0 specification (fka Swagger 2.0 specification) from annotated go code.
      
> [!NOTE]
> Inherited *key aspects* from baseline code:
>
> - **Hexagonal architecture**
> 
> - **uber-go/fx** dependency injection framework
> 
> - **uber-go/config** injection-friendly YAML configuration
> 
> - **net/http** with **CORS** security
> 
> - **gorilla/mux** router and a **subrouter** to use specific middleware for specific routes
> 
> - **rs/zerolog** logging library. **Per request contextual logging**  (all traces within the same request will share the same unique id)
>
> - **OpenId Connect** token security

## Changes to baseline code


## Support, Questions, or Feedback

I'll accept pretty much everything so feel free to open a Pull-Request
