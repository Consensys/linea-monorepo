import { ApolloClient, HttpLink, InMemoryCache, from } from "@apollo/client";
import { SetContextLink } from "@apollo/client/link/context";
import { IOAuth2TokenClient } from "@consensys/linea-shared-utils";

// Could move to linea-shared-utils one day, if there is another project using @apollo/client
export function createApolloClient(oAuth2TokenClient: IOAuth2TokenClient, graphqlUri: string): ApolloClient {
  // --- create the base HTTP transport
  const httpLink = new HttpLink({
    uri: graphqlUri,
  });

  const asyncAuthLink = new SetContextLink(async (prevContext, operation) => {
    void operation;
    const bearerToken = await oAuth2TokenClient.getBearerToken();
    return {
      headers: {
        ...prevContext.headers,
        authorization: bearerToken,
      },
    };
  });
  // --- combine links so authLink runs before httpLink
  const client = new ApolloClient({
    link: from([asyncAuthLink, httpLink]),
    cache: new InMemoryCache(),
  });
  return client;
}
