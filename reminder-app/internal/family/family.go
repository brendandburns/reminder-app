package family

type Family struct {
    ID      string   `json:"id"`
    Name    string   `json:"name"`
    Members []string `json:"members"`
}

func (f *Family) AddMember(member string) {
    f.Members = append(f.Members, member)
}

func (f *Family) RemoveMember(member string) {
    for i, m := range f.Members {
        if m == member {
            f.Members = append(f.Members[:i], f.Members[i+1:]...)
            break
        }
    }
}

func (f *Family) GetMembers() []string {
    return f.Members
}