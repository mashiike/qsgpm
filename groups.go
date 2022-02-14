package qsgpm

type Groups map[string]Group

func newGroups() Groups {
	return make(Groups)
}

type Group struct {
	name       string
	membership map[string]struct{}
}

func (groups Groups) AddGroup(group string) Group {
	g, ok := groups[group]
	if !ok {
		g = Group{
			name:       group,
			membership: make(map[string]struct{}),
		}
		groups[group] = g
	}
	return g
}
func (groups Groups) Add(group, user string) {
	g := groups.AddGroup(group)
	g.Add(user)
}

func (groups Groups) Assgin(user string, groupNames []string) {
	for _, g := range groupNames {
		groups.Add(g, user)
	}
}

func (groups Groups) DiffGroup(other Groups) (add, stable, delete []string) {
	add = make([]string, 0)
	delete = make([]string, 0)
	stable = make([]string, 0)
	for o := range other {
		if _, ok := groups[o]; ok {
			stable = append(stable, o)
		} else {
			add = append(add, o)
		}
	}
	for g := range groups {
		if _, ok := other[g]; !ok {
			delete = append(delete, g)
		}
	}
	return
}

type Membership struct {
	GroupName string
	UserName  string
}

func (groups Groups) DiffMembership(other Groups) (add, stable, delete []Membership) {
	addGroups, stableGroups, deleteGroups := groups.DiffGroup(other)
	add = make([]Membership, 0)
	delete = make([]Membership, 0)
	stable = make([]Membership, 0)
	for _, addGroup := range addGroups {
		add = append(add, other[addGroup].Membership()...)
	}
	for _, deleteGroup := range deleteGroups {
		delete = append(delete, groups[deleteGroup].Membership()...)
	}
	for _, stableGroup := range stableGroups {
		addMembership, stableMembership, deleteMembership := groups[stableGroup].Diff(other[stableGroup])
		add = append(add, addMembership...)
		delete = append(delete, deleteMembership...)
		stable = append(stable, stableMembership...)
	}
	return
}

func (group Group) Add(user string) {
	group.membership[user] = struct{}{}
}

func (group Group) Membership() []Membership {
	membership := make([]Membership, 0, len(group.membership))
	for user := range group.membership {
		membership = append(membership, Membership{
			GroupName: group.name,
			UserName:  user,
		})
	}
	return membership
}

func (group Group) Diff(other Group) (add, stable, delete []Membership) {
	if group.name != other.name {
		panic("unexpected diff operation")
	}
	add = make([]Membership, 0)
	delete = make([]Membership, 0)
	stable = make([]Membership, 0)
	for o := range other.membership {
		if _, ok := group.membership[o]; ok {
			stable = append(stable, Membership{
				GroupName: group.name,
				UserName:  o,
			})
		} else {
			add = append(add, Membership{
				GroupName: group.name,
				UserName:  o,
			})
		}
	}
	for g := range group.membership {
		if _, ok := other.membership[g]; !ok {
			delete = append(delete, Membership{
				GroupName: group.name,
				UserName:  g,
			})
		}
	}
	return
}
