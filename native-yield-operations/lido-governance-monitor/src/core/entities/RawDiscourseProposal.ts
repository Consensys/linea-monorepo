export interface RawDiscoursePost {
  id: number;
  username: string;
  cooked: string;
  post_url: string;
  created_at: string;
}

export interface RawDiscourseProposal {
  id: number;
  slug: string;
  title: string;
  created_at: string;
  post_stream: {
    posts: RawDiscoursePost[];
  };
}

export interface RawDiscourseProposalList {
  topic_list: {
    topics: Array<{
      id: number;
      slug: string;
    }>;
  };
}
